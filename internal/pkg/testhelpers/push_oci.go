package testhelpers

import (
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"

	"dagger/cuestomize/shared/oci"
)

func PushDirectoryToOCIRegistry_T(t *testing.T, reference, rootDirectory, artifactType, tag string, client remote.Client, plainHTTP bool) ocispec.Descriptor {
	t.Helper()

	descriptor, err := oci.PushDirectoryToOCIRegistry(t.Context(), reference, rootDirectory, artifactType, tag, client, plainHTTP)
	if err != nil {
		t.Fatalf("Failed to push directory to OCI registry: %v", err)
	}

	return descriptor
}
