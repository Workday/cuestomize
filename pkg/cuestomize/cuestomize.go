package cuestomize

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/cuerrors"
	"github.com/go-logr/logr"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// Cuestomize generates (or validates) resources from the provided CUE configuration and input resources.
func Cuestomize(ctx context.Context, items []*kyaml.RNode, config *api.KRMInput, opts ...Option) ([]*kyaml.RNode, error) {
	log := logr.FromContextOrDiscard(ctx)
	detailer := cuerrors.FromContextOrEmpty(ctx)

	var cuestomizeOpts options
	for _, opt := range opts {
		opt(&cuestomizeOpts)
	}

	if err := cuestomizeOpts.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	err := cuestomizeOpts.ModelProvider.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CUE model from provider: %w", err)
	}

	resourcesPath := cuestomizeOpts.ModelProvider.Path()

	cueCtx := cuecontext.New()

	includes, err := api.ExtractIncludes(ctx, config, items)
	if err != nil {
		return nil, fmt.Errorf("failed to compute includes from KRM function inputs: %w", err)
	}
	includesValue, err := includes.IntoCueValue(cueCtx)
	if err != nil {
		return nil, detailer.ErrorWithDetails(err, "failed to convert includes into CUE value")
	}

	configValue, err := config.IntoCueValue(cueCtx)
	if err != nil {
		return nil, detailer.ErrorWithDetails(err, "failed to convert config into CUE value")
	}

	instances, err := LoadCUEModel(ctx, resourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CUE model from '%s': %w", resourcesPath, err)
	}

	schema, err := BuildCUEModelSchema(ctx, cueCtx, instances)
	if err != nil {
		return nil, fmt.Errorf("failed to build CUE model schema: %w", err)
	}

	unified, err := FillMetadata(ctx, *schema, config)
	if err != nil {
		return nil, fmt.Errorf("failed to fill metadata in CUE schema: %w", err)
	}
	unified = unified.FillPath(cue.ParsePath(InputFillPath), configValue)
	unified = unified.FillPath(cue.ParsePath(IncludesFillPath), includesValue)
	if unified.Err() != nil {
		return nil, detailer.ErrorWithDetails(unified.Err(), "failed to unify CUE model with inputs from KRM function")
	}

	// assert that the unified instance values are all concrete (no string, regexes, etc.)
	// without this check, non-valorised fields can remain in output resources
	if err := unified.Validate(cue.Final(), cue.Concrete(true)); err != nil {
		return nil, detailer.ErrorWithDetails(err, "failed to validate unified CUE instance")
	}

	if ShouldActAsValidator(config) {
		log.V(4).Info("cuestomize is acting in validator mode.")
		return items, nil // if the function is a validator, return the original items without processing
	}
	return ProcessOutputs(ctx, unified, items)
}
