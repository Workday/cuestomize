// Package fetcher provides functionality for fetching artifacts from OCI registries.
package fetcher

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// FetchFromOCIRegistry fetches an artifact from an OCI registry and stores it in the specified working directory.
func FetchFromOCIRegistry(ctx context.Context, client remote.Client, workingDir string, ref registry.Reference, plainHTTP bool) error {
	log := logr.FromContextOrDiscard(ctx).V(4)

	fs, err := file.New(workingDir)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}

	repository, err := remote.NewRepository(ref.String())
	if err != nil {
		return err
	}
	if client != nil {
		repository.Client = client
	}
	repository.PlainHTTP = plainHTTP

	desc, err := oras.Copy(ctx, repository, ref.Reference, fs, ref.Reference, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	log.Info("fetched artifact from OCI registry",
		"reg", ref.Registry,
		"repo", ref.Repository,
		"tag", ref.Reference,
		"workingDir", workingDir,
		"digest", desc.Digest.String(),
		"mediaType", desc.MediaType,
	)

	return nil
}
