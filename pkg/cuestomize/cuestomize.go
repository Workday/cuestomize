package cuestomize

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/internal/pkg/cuerrors"
	"github.com/Workday/cuestomize/internal/pkg/cuestomize/oci"
	"github.com/go-logr/logr"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// Cuestomize generates (or validates) resources from the provided CUE configuration and input resources.
func Cuestomize(items []*kyaml.RNode, config *api.KRMInput, resourcesPath string, ctx context.Context) ([]*kyaml.RNode, error) {
	log := logr.FromContextOrDiscard(ctx)

	detailer := cuerrors.NewDetailerWithCwd(resourcesPath)
	ctx = cuerrors.NewContext(ctx, detailer)

	cueCtx := cuecontext.New()

	if config.RemoteModule != nil {
		log.V(4).Info("fetching CUE model from OCI registry")
		if err := oci.FetchFromRegistry(ctx, config, items, resourcesPath); err != nil {
			return nil, fmt.Errorf("failed to fetch from OCI registry: %w", err)
		}
	}

	includes, err := api.ExtractIncludes(config, items, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compute includes from KRM function inputs: %w", err)
	}
	includesValue, err := includes.IntoCueValue(ctx, cueCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert includes into CUE value: %w", err)
	}

	configValue, err := config.IntoCueValue(ctx, cueCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config into CUE value: %w", err)
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
