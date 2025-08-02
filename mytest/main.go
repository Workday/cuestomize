package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
)

// PushDirectoryToOCIRegistry walks a local directory, packs its contents into an
// OCI artifact, and pushes it to a remote repository.
func PushDirectoryToOCIRegistry(ctx context.Context, reference, rootDirectory, artifactType string) (ocispec.Descriptor, error) {
	// 1. Set up a connection to the remote repository.
	repo, err := remote.NewRepository(reference)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to create repository: %w", err)
	}
	repo.PlainHTTP = true // Use plain HTTP for local testing; set to false for production

	// 2. Create a file store and gather file descriptors from the directory.
	// Using file.New("") creates an in-memory store that we'll populate.
	fileStore, err := file.New("")
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to create file store: %w", err)
	}
	defer fileStore.Close()

	fileDescriptors := []ocispec.Descriptor{}

	// Walk the specified directory to find all files.
	err = filepath.WalkDir(rootDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// // Skip the .oras metadata directory itself.
		// if d.IsDir() && d.Name() == ".oras" {
		// 	return filepath.SkipDir
		// }

		// Skip directories, as we only want to add files.
		if !d.IsDir() {
			// Use the path relative to the root directory as the name of the file in the artifact.
			// This preserves the directory structure.
			nameInArtifact, err := filepath.Rel(rootDirectory, path)
			if err != nil {
				return err
			}

			// Add the file to the in-memory store. The `path` is the file's location
			// on disk, and `nameInArtifact` is how it will be identified in the manifest.
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

	// 3. Pack all the file descriptors into a single OCI manifest.
	// This manifest will have a layer for each file in your directory.
	// func oras.PackManifest(ctx context.Context, pusher content.Pusher, packManifestVersion oras.PackManifestVersion, artifactType string, opts oras.PackManifestOptions) (ocispec.Descriptor, error)
	manifestDescriptor, err := oras.PackManifest(ctx, fileStore, oras.PackManifestVersion1_1, artifactType, oras.PackManifestOptions{
		Layers: fileDescriptors,
	})
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to pack artifact: %w", err)
	}

	tag := "latest"
	if err = fileStore.Tag(ctx, manifestDescriptor, tag); err != nil {
		panic(err)
	}

	// 4. Push the artifact (manifest and all file blobs) to the remote repository.
	pushedDescriptor, err := oras.Copy(ctx, fileStore, tag, repo, reference, oras.DefaultCopyOptions)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to push artifact: %w", err)
	}

	return pushedDescriptor, nil
}

func main() {
	// --- Configuration ---
	// ❗ IMPORTANT: Change this to your OCI registry reference.
	// To run a local one: docker run -d -p 5000:5000 --restart=always --name registry registry:2
	registryRef := "localhost:5001/cue-test-my-directory-artifact:latest"
	dirToPush := "my_app_bundle"
	artifactType := "application/vnd.example.directory.v1"

	// --- Setup a dummy directory with files to push ---
	if err := os.Mkdir(dirToPush, 0755); err != nil {
		log.Printf("Failed to create root directory: %v", err)
		return
	}
	defer os.RemoveAll(dirToPush) // Clean up the directory and its contents
	// if err := os.Mkdir(".oras", 0755); err != nil {
	// 	log.Fatalf("Failed to create .oras directory: %v", err)
	// }

	// Create some files inside the directory
	_ = os.WriteFile(filepath.Join(dirToPush, "config.yaml"), []byte("api_version: v1\n"), 0644)
	_ = os.Mkdir(filepath.Join(dirToPush, "bin"), 0755)
	_ = os.WriteFile(filepath.Join(dirToPush, "bin", "app.exe"), []byte{0xDE, 0xAD, 0xBE, 0xEF}, 0644)
	_ = os.WriteFile(filepath.Join(dirToPush, "README.md"), []byte("# My App Bundle"), 0644)

	fmt.Printf("🚀 Pushing directory '%s/' to '%s'...\n", dirToPush, registryRef)

	// --- Call the push function ---
	ctx := context.Background()
	desc, err := PushDirectoryToOCIRegistry(ctx, registryRef, dirToPush, artifactType)
	if err != nil {
		log.Printf("❌ Push failed: %v", err)
		return
	}

	fmt.Println("\n✅ Successfully pushed artifact!")
	fmt.Printf("Digest: %s\n", desc.Digest)
	fmt.Printf("Size: %d bytes\n", desc.Size)

	fmt.Println("\nTo verify the contents, you can pull the artifact with the `oras` CLI:")
	fmt.Printf("oras pull %s\n", registryRef)
	fmt.Printf("Then inspect the downloaded '%s' directory.\n", dirToPush)
}
