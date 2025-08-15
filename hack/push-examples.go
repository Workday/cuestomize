package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"dagger/cuestomize/shared/oci"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	CueModuleArtifactType     = "application/vnd.cue.module.v1+json"
	CueModuleFileArtifactType = "application/vnd.cue.modulefile.v1"

	LogLevelEnvVar = "LOG_LEVEL"
)

func main() {
	ctx := context.Background()
	if err := setupLogging(); err != nil {
		panic(fmt.Errorf("failed to setup logging: %w", err))
	}
	log.Debug().Msg("Starting to push examples to OCI registry")
	username := os.Getenv("OCI_USERNAME")
	if username == "" {
		panic("OCI_USERNAME environment variable is not set")
	}
	password := os.Getenv("OCI_PASSWORD")
	if password == "" {
		panic("OCI_PASSWORD environment variable is not set")
	}
	registry := os.Getenv("OCI_REGISTRY")
	if registry == "" {
		panic("OCI_REGISTRY environment variable is not set")
	}
	repositoryPrefix := os.Getenv("OCI_REPOSITORY_PREFIX")
	if repositoryPrefix == "" {
		panic("OCI_REPOSITORY_PREFIX environment variable is not set")
	}
	latest := os.Getenv("IS_LATEST") == "true"
	if len(os.Args) < 2 {
		panic("pass tag as argument")
	}
	// first arg
	examplesDir := "./examples"
	tag := os.Args[1]

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		panic(fmt.Errorf("failed to read examples directory: %w", err))
	}
	log.Trace().Int("entries", len(entries)).Msg("Found entries in examples directory")

	repositoryDirMap := make(map[string]string)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(path.Join(examplesDir, entry.Name(), "cue", "cue.mod")); err != nil {
			log.Warn().Err(err).Str("entry", entry.Name()).Msg("Skipping example without cue/cue.mod file")
			continue
		}
		log.Trace().Str("entry", entry.Name()).Msg("Found example in directory")
		repoName := registry + "/" + repositoryPrefix + "-" + entry.Name()
		repositoryDirMap[repoName] = path.Join(examplesDir, entry.Name())
	}

	tags := []string{tag}
	if latest {
		tags = append(tags, "latest")
	}
	log.Trace().Strs("tags", tags).Msg("Tags to push")

	client := auth.DefaultClient

	client.Credential = auth.StaticCredential(registry, auth.Credential{
		Username: username,
		Password: password,
	})

	pushedRefs := make([]string, 0, len(repositoryDirMap)*len(tags))
	for repoName, dir := range repositoryDirMap {
		for _, t := range tags {
			repoWithTag := repoName + ":" + t
			log.Debug().Str("repoWithTag", repoWithTag).Msg("Pushing to OCI registry")
			_, err := oci.PushDirectoryToOCIRegistry(ctx, repoWithTag, dir, CueModuleArtifactType, t, client, false)
			log.Info().Str("repoWithTag", repoWithTag).Msg("Pushed to OCI registry")
			if err != nil {
				panic(fmt.Errorf("failed to push %s to OCI registry: %w", repoWithTag, err))
			}
			pushedRefs = append(pushedRefs, repoWithTag)
		}
	}
	log.Info().Strs("pushedRefs", pushedRefs).Msg("Pushed references to OCI registry")
}

// setupLogging configures the global logging level based on the log level environment variable.
func setupLogging() error {
	logLevel := os.Getenv(LogLevelEnvVar)
	if logLevel == "" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level from environment variable %s: %w", LogLevelEnvVar, err)
	}
	zerolog.SetGlobalLevel(level)
	return nil
}
