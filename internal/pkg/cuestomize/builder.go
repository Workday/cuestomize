// Package cuestomize provides a builder for creating KRM (Kubernetes Resource Model) functions
// that generate Kubernetes resources from CUE configurations. It enables users to define
// resource transformations using CUE language and apply them as part of a Kustomize pipeline.
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

// KRMFuncBuilder is a builder for the Cuestomize KRM function.
type KRMFuncBuilder struct {
	// resourcesPath is the path to the directory containing the CUE resources.
	// If not set, it defaults to "/cue-resources".
	resourcesPath string

	// config is a pointer to the configuration object the KRM function will receive in input.
	config *api.KRMInput
}

// NewBuilder creates a new KRMFuncBuilder with the default resources path.
func NewBuilder() *KRMFuncBuilder {
	return &KRMFuncBuilder{
		resourcesPath: DefaultResourcesPath,
	}
}

// SetConfig sets the reference to the configuration object that the KRM function will receive in input.
func (b *KRMFuncBuilder) SetConfig(config *api.KRMInput) *KRMFuncBuilder {
	b.config = config
	return b
}

// SetResourcesPath sets the path to the directory containing the CUE resources.
func (b *KRMFuncBuilder) SetResourcesPath(resourcesPath string) *KRMFuncBuilder {
	b.resourcesPath = resourcesPath
	return b
}

// Build returns a function that can be used to generate resources from a CUE configuration and some input resources.
func (b *KRMFuncBuilder) Build(ctx context.Context) (func([]*kyaml.RNode) ([]*kyaml.RNode, error), error) {
	if b.config == nil {
		return nil, fmt.Errorf("config must be set before building the KRM function")
	}
	return newCuestomizeFunctionWithPath(ctx, b.config, &b.resourcesPath), nil
}
