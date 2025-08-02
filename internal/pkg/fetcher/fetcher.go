package fetcher

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
)

// FetchFromOCIRegistry fetches an artifact from an OCI registry and stores it in the specified working directory.
func FetchFromOCIRegistry(ctx context.Context, client *remote.Client, workingDir, reg, repo, tag string, plainHTTP bool) error {
	fs, err := file.New(workingDir)
	if err != nil {
		log.Error().Err(err).Str("workingDir", workingDir).Msg("failed to create file store")
		return fmt.Errorf("failed to create file store: %w", err)
	}

	repository, err := remote.NewRepository(reg + "/" + repo)
	if err != nil {
		log.Error().Err(err).Str("reg", reg).Str("repo", repo).Msg("failed to create remote repository")
		return err
	}
	if client != nil {
		repository.Client = *client
	}
	repository.PlainHTTP = plainHTTP

	desc, err := oras.Copy(ctx, repository, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	log.Debug().Str("reg", reg).
		Str("repo", repo).
		Str("workingDir", workingDir).
		Str("digest", desc.Digest.String()).
		Str("mediaType", desc.MediaType).
		Msg("fetched artifact from OCI registry")
	return nil
}
