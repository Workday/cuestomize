// Package testhelpers provides utility functions for to be used in tests.
package testhelpers

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	yaml "sigs.k8s.io/yaml"
)

// LoadFromFile is a helper function to read the specified file and unmarshal it into the provided type T.
func LoadFromFile[T interface{}](t *testing.T, path string) *T {
	t.Helper()

	krmInputFile, err := os.ReadFile(path)
	assert.NoError(t, err, "failed to read %s", path)

	krmInput := new(T)
	err = yaml.Unmarshal(krmInputFile, krmInput)
	assert.NoError(t, err, "failed to unmarshal %s into KRMInput", path)

	return krmInput
}

// LoadResourceList is a helper function to read the items file from the specified directory.
// It returns a slice of RNodes representing the input and the items (as in an actual run of a KRM function).
func LoadResourceList(t *testing.T, krmInputPath, itemsPath string) []*kyaml.RNode {
	t.Helper()

	// Read the KRMInput file and append it to the resource list.
	krmInput, err := kyaml.ReadFile(krmInputPath)
	assert.NoError(t, err, "failed to read %s", krmInputPath)

	// Read the items file and append its resources to the resource list.
	itemsFile, err := os.ReadFile(itemsPath)
	assert.NoError(t, err, "failed to read %s", itemsPath)

	resources, err := (&kio.ByteReader{Reader: bytes.NewReader(itemsFile)}).Read()
	assert.NoError(t, err, "failed to read items from %s", itemsPath)

	// create the resource list
	resourceList := make([]*kyaml.RNode, 0, len(resources)+1)
	resourceList = append(resourceList, krmInput)
	resourceList = append(resourceList, resources...)

	return resourceList
}
