package oci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/internal/pkg/fetcher"
	"github.com/go-logr/logr"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// FetchFromRegistry fetches a CUE module from the remote defined in the config, and places it in the working directory.
func FetchFromRegistry(ctx context.Context, config *api.KRMInput, items []*kyaml.RNode, workingDir string) error {
	log := logr.FromContextOrDiscard(ctx).V(2)

	client, err := config.GetRemoteClient(items)
	if err != nil {
		return fmt.Errorf("failed to configure remote client: %w", err)
	}

	log.Info("fetching from OCI registry",
		"registry", config.RemoteModule.Registry,
		"repo", config.RemoteModule.Repo,
		"tag", config.RemoteModule.Tag,
		"plainHTTP", config.RemoteModule.PlainHTTP,
	)

	if err := fetcher.FetchFromOCIRegistry(
		ctx,
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
		log.V(-1).Info("cue.mod directory not found in artifact. This might cause Cuestomize issues interacting with the module.", "error", err)
	}

	return nil
}
