// Package main provides a utility to push CUE module examples to an OCI registry.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/Workday/cuestomize/pkg/oci"
	"github.com/go-logr/logr"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	CueModuleArtifactType     = "application/vnd.cue.module.v1+json"
	CueModuleFileArtifactType = "application/vnd.cue.modulefile.v1"

	LogLevelEnvVar = "LOG_LEVEL"
)

func main() {
	ctx, err := setupLogging(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to setup logging: %w", err))
	}
	log := logr.FromContextOrDiscard(ctx)

	log.V(4).Info("starting to push examples to OCI registry")
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
	repositoryPrefix = strings.ToLower(repositoryPrefix)
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

	log.V(4).Info("found entries in examples directory", "entries", len(entries))

	repositoryDirMap := make(map[string]string)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(path.Join(examplesDir, entry.Name(), "cue", "cue.mod")); err != nil {
			log.Info("skipping example without cue/cue.mod file", "entry", entry.Name(), "error", err)
			continue
		}
		log.V(4).Info("found example in directory", "entry", entry.Name())
		repoName := registry + "/" + repositoryPrefix + "-" + entry.Name()
		repositoryDirMap[repoName] = path.Join(examplesDir, entry.Name(), "cue")
	}

	tags := []string{tag}
	if latest {
		tags = append(tags, "latest")
	}
	log.V(4).Info("tags to push", "tags", tags)

	client := auth.DefaultClient

	client.Credential = auth.StaticCredential(registry, auth.Credential{
		Username: username,
		Password: password,
	})

	pushedRefs := make([]string, 0, len(repositoryDirMap)*len(tags))
	for repoName, dir := range repositoryDirMap {
		for _, t := range tags {
			repoWithTag := repoName + ":" + t
			log.V(4).Info("pushing to OCI registry", "repoWithTag", repoWithTag, "dir", dir)
			_, err := oci.PushDirectoryToOCIRegistry(ctx, repoWithTag, dir, CueModuleArtifactType, t, client, false)
			if err != nil {
				panic(fmt.Errorf("failed to push %s to OCI registry: %w", repoWithTag, err))
			}
			log.Info("pushed to OCI registry", "repoWithTag", repoWithTag)
			pushedRefs = append(pushedRefs, repoWithTag)
		}
	}
	log.Info("pushed references to OCI registry", "pushedRefs", pushedRefs)
}

// setupLogging configures the global logging level based on the log level environment variable.
func setupLogging(ctx context.Context) (context.Context, error) {
	logLevel := os.Getenv(LogLevelEnvVar)

	lvl := slog.LevelInfo
	if logLevel != "" {
		err := lvl.UnmarshalText([]byte(logLevel))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal log level from environment variable %s: %w", LogLevelEnvVar, err)
		}
	}

	log := logr.FromSlogHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))

	return logr.NewContext(ctx, log), nil
}
