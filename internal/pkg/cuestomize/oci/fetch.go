package oci

import (
	"context"
	"fmt"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/oci/fetcher"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// FetchFromRegistry fetches a CUE module from the remote defined in the config, and places it in the working directory.
func FetchFromRegistry(ctx context.Context, config *api.KRMInput, items []*kyaml.RNode, workingDir string) error {
	client, err := config.GetRemoteClient(items)
	if err != nil {
		return fmt.Errorf("failed to configure remote client: %w", err)
	}

	log.Debug().Str("registry", config.RemoteModule.Registry).
		Str("repo", config.RemoteModule.Repo).
		Str("tag", config.RemoteModule.Tag).
		Bool("plainHTTP", config.RemoteModule.PlainHTTP).
		Msg("fetching from OCI registry")

	if err := fetcher.FetchFromOCIRegistry(
		context.TODO(),
		client,
		workingDir,
		config.RemoteModule.Registry,
		config.RemoteModule.Repo,
		config.RemoteModule.Tag,
		config.RemoteModule.PlainHTTP,
	); err != nil {
		return fmt.Errorf("failed to fetch from OCI registry: %w", err)
	}
	_, err = os.Stat(filepath.Join(workingDir, "cue.mod"))
	if err != nil {
		log.Warn().Err(err).Msg("cue.mod directory not found. Please ensure cue.mod exists in artifact")
	}

	return nil
}
