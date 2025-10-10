package testhelpers

import (
	"testing"

	"github.com/Workday/cuestomize/pkg/oci"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"
)

// PushDirectoryToOCIRegistryT is a test helper that pushes the contents of a directory to an OCI registry.
func PushDirectoryToOCIRegistryT(t *testing.T, reference, rootDirectory, artifactType, tag string, client remote.Client, plainHTTP bool) ocispec.Descriptor {
	t.Helper()

	descriptor, err := oci.PushDirectoryToOCIRegistry(t.Context(), reference, rootDirectory, artifactType, tag, client, plainHTTP)
	if err != nil {
		t.Fatalf("Failed to push directory to OCI registry: %v", err)
	}

	return descriptor
}
