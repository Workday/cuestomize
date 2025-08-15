package oci

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
)

// PushDirectoryToOCIRegistry walks a local directory, packs its contents into an
// OCI artifact, and pushes it to a remote repository.
func PushDirectoryToOCIRegistry(ctx context.Context, reference, rootDirectory, artifactType, tag string, client remote.Client, plainHTTP bool) (ocispec.Descriptor, error) {
	repo, err := remote.NewRepository(reference)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to create repository: %w", err)
	}
	if client != nil {
		repo.Client = client
	}
	repo.PlainHTTP = plainHTTP

	// creates an in-memory store
	fileStore, err := file.New("")
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to create file store: %w", err)
	}
	defer fileStore.Close()

	fileDescriptors := []ocispec.Descriptor{}

	err = filepath.WalkDir(rootDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, as we only want to add files.
		if !d.IsDir() {
			// use the path relative to the root directory as the name of the file in the artifact.
			nameInArtifact, err := filepath.Rel(rootDirectory, path)
			if err != nil {
				return err
			}

			fileDescriptor, err := fileStore.Add(ctx, nameInArtifact, "", path)
			if err != nil {
				return fmt.Errorf("failed to add file %q to store: %w", path, err)
			}
			fileDescriptors = append(fileDescriptors, fileDescriptor)
		}
		return nil
	})

	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to walk directory %q: %w", rootDirectory, err)
	}
	if len(fileDescriptors) == 0 {
		return ocispec.Descriptor{}, fmt.Errorf("no files found in directory %q", rootDirectory)
	}

	// pack all the file descriptors into a single OCI manifest.
	// This manifest will have a layer for each file in your directory.
	manifestDescriptor, err := oras.PackManifest(ctx, fileStore, oras.PackManifestVersion1_1, artifactType, oras.PackManifestOptions{
		Layers: fileDescriptors,
	})
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to pack artifact: %w", err)
	}

	if err = fileStore.Tag(ctx, manifestDescriptor, tag); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to tag artifact: %w", err)
	}

	// push the artifact (manifest and all file blobs) to the remote repository
	pushedDescriptor, err := oras.Copy(ctx, fileStore, tag, repo, reference, oras.DefaultCopyOptions)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to push artifact: %w", err)
	}

	return pushedDescriptor, nil
}
