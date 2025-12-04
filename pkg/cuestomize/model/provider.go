package model

import (
	"context"
)

// Provider defines the interface for a CUE model provider.
//
// A Provider is responsible for making the CUE model available at a specific path.
//
// The simplest implementation is a local file system provider that loads the CUE model from a local directory.
type Provider interface {
	// Get ensures that the CUE model is available at the specified path.
	Get(ctx context.Context) error
	// Path returns the file system path where the CUE model is located.
	Path() string
}
