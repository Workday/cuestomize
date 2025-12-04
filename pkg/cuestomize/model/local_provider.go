// Package model provides abstractions for CUE model fetching, and default implementations for local and OCI-based model fetching.
package model

import (
	"context"
)

// LocalPathProvider is a model provider that uses a local file system path.
type LocalPathProvider struct {
	resourcesPath string
}

// Path returns the local file system path to the CUE model.
func (p *LocalPathProvider) Path() string {
	return p.resourcesPath
}

// NewLocalPathProvider creates a new LocalPathProvider with the given resources path.
func NewLocalPathProvider(resourcesPath string) *LocalPathProvider {
	return &LocalPathProvider{resourcesPath: resourcesPath}
}

// Get is a no-op for LocalPathProvider since the model is already available locally.
func (p *LocalPathProvider) Get(_ context.Context) error {
	return nil
}
