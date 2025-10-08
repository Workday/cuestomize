package cuestomize

import (
	"context"

	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/cuestomize"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// KRMFunction is the function type of Kustomize KRM functions.
type KRMFunction = func([]*kyaml.RNode) ([]*kyaml.RNode, error)

// newCuestomizeFunctionWithPath returns a function that can be used to generate resources
// from a CUE configuration and input resources.
//
// Input:
//
// * config: pointer to the configuration object
//
// * resourcesPath: path to the directory containing the CUE resources (nil to use the default)
func newCuestomizeFunctionWithPath(config *api.KRMInput, resourcesPath *string, ctx context.Context) KRMFunction {
	return func(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
		return cuestomize.Cuestomize(items, config, *resourcesPath, ctx)
	}
}
