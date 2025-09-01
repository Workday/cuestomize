package oci

import (
	"context"
	"fmt"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/internal/pkg/fetcher"
	"github.com/rs/zerolog/log"
	"os"
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
	entries, err := os.ReadDir("./")
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}

	for _, e := range entries {
		os.Stderr.WriteString(e.Name())
	}
	return nil
}
