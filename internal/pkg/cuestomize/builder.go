package cuestomize

import (
	"context"
	"fmt"

	"github.com/Workday/cuestomize/api"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	// DefaultResourcesPath is the default path to the directory containing the CUE resources.
	DefaultResourcesPath = "/cue-resources"
)

// CUEstomizeFuncBuilder is a builder for the CUEstomize KRM function.
type CUEstomizeFuncBuilder struct {
	// resourcesPath is the path to the directory containing the CUE resources.
	// If not set, it defaults to "/cue-resources".
	resourcesPath string

	// config is a pointer to the configuration object the KRM function will receive in input.
	config *api.KRMInput
}

// NewBuilder creates a new CUEstomizeFuncBuilder with the default resources path.
func NewBuilder() *CUEstomizeFuncBuilder {
	return &CUEstomizeFuncBuilder{
		resourcesPath: DefaultResourcesPath,
	}
}

// SetConfig sets the reference to the configuration object that the KRM function will receive in input.
func (b *CUEstomizeFuncBuilder) SetConfig(config *api.KRMInput) *CUEstomizeFuncBuilder {
	b.config = config
	return b
}

// SetResourcesPath sets the path to the directory containing the CUE resources.
func (b *CUEstomizeFuncBuilder) SetResourcesPath(resourcesPath string) *CUEstomizeFuncBuilder {
	b.resourcesPath = resourcesPath
	return b
}

// Build returns a function that can be used to generate resources from a CUE configuration and some input resources.
func (b *CUEstomizeFuncBuilder) Build(ctx context.Context) (func([]*kyaml.RNode) ([]*kyaml.RNode, error), error) {
	if b.config == nil {
		return nil, fmt.Errorf("config must be set before building the KRM function")
	}
	return newCuestomizeFunctionWithPath(b.config, &b.resourcesPath, ctx), nil
}
