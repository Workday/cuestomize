package cuestomize

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/internal/pkg/cuerrors"
	"github.com/Workday/cuestomize/internal/pkg/cuestomize/oci"
	"github.com/rs/zerolog/log"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	// InputFillPath is the CUE path in which the input resources will be injected into the CUE model.
	InputFillPath = "input"
	// IncludesFillPath is the CUE path in which the includes will be injected into the CUE model.
	IncludesFillPath = "includes"

	// OutputsPath is the CUE path in which the function expects the output resources (as a list) to be placed.
	OutputsPath = "outputs"
)

const (
	// APIVersionFillPath is the CUE path in which the API version of the KRMInput will be filled.
	APIVersionFillPath = "apiVersion"
	// KindFillPath is the CUE path in which the kind of the KRMInput will be filled.
	KindFillPath = "kind"
	// MetadataFillPath is the CUE path in which the metadata of the KRMInput will be filled.
	MetadataFillPath = "metadata"
)

const (
	// ValidatorAnnotationKey is the annotation key that marks a CUE function as a validator.
	ValidatorAnnotationKey = "config.cuestomize.io/validator"
	// ValidatorAnnotationValue is the value of the annotation that marks a CUE function as a validator.
	ValidatorAnnotationValue = "true"
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
func newCuestomizeFunctionWithPath(config *api.KRMInput, resourcesPath *string) KRMFunction {
	return func(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
		ctx := cuecontext.New()

		if config.RemoteModule != nil {
			log.Debug().Msg("fetching CUE model from OCI registry")
			if err := oci.FetchFromRegistry(context.TODO(), config, items, *resourcesPath); err != nil {
				return nil, fmt.Errorf("failed to fetch from OCI registry: %w", err)
			}
		}

		includes, err := api.ExtractIncludes(config, items)
		if err != nil {
			return nil, fmt.Errorf("failed to compute includes from KRM function inputs: %w", err)
		}
		includesValue, err := includes.IntoCueValue(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert includes into CUE value: %w", err)
		}

		configValue, err := config.IntoCueValue(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert config into CUE value: %w", err)
		}

		instances, err := loadCUEModel(*resourcesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load CUE model from '%s': %w", *resourcesPath, err)
		}

		schema, err := buildCUEModelSchema(ctx, instances)
		if err != nil {
			return nil, fmt.Errorf("failed to build CUE model schema: %w", err)
		}

		unified, err := fillMetadata(*schema, config)
		if err != nil {
			return nil, fmt.Errorf("failed to fill metadata in CUE schema: %w", err)
		}
		unified = unified.FillPath(cue.ParsePath(InputFillPath), configValue)
		unified = unified.FillPath(cue.ParsePath(IncludesFillPath), includesValue)
		if unified.Err() != nil {
			return nil, cuerrors.ErrorWithDetails(unified.Err(), "failed to unify CUE model with inputs from KRM function")
		}

		// assert that the unified instance values are all concrete (no string, regexes, etc.)
		// without this check, non-valorised fields can remain in output resources
		if err := unified.Validate(cue.Final(), cue.Concrete(true)); err != nil {
			return nil, cuerrors.ErrorWithDetails(err, "failed to validate unified CUE instance")
		}

		if shouldActAsValidator(config) {
			log.Debug().Msg("function is acting in validator mode.")
			return items, nil // if the function is a validator, return the original items without processing
		}
		return processOutputs(unified, items)
	}
}

// buildCUEModelSchema builds a CUE model from the provided instances and returns the unified schema.
func buildCUEModelSchema(ctx *cue.Context, instances []*build.Instance) (*cue.Value, error) {
	values, err := ctx.BuildInstances(instances)
	if err != nil {
		return nil, fmt.Errorf("failed to build CUE instances: %w", err)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("no CUE values found after building instances")
	}

	schema := values[0]
	for i := 1; i < len(values); i++ {
		schema = schema.Unify(values[i])
		if schema.Err() != nil {
			return nil, cuerrors.ErrorWithDetails(schema.Err(), "failed to unify CUE model with [%v]", instances[i].BuildFiles)
		}
	}

	return &schema, nil
}

// loadCUEModel loads a CUE model from the specified path and returns the instances.
func loadCUEModel(path string) ([]*build.Instance, error) {
	cfg := &load.Config{Dir: path}
	instances := load.Instances([]string{"."}, cfg)
	if len(instances) == 0 {
		return nil, fmt.Errorf("no CUE instances found")
	}

	return instances, checkInstances(instances)
}

// checkInstances checks if any of the instances have an error and returns an error if so.
func checkInstances(instances []*build.Instance) error {
	for _, inst := range instances {
		if inst.Err != nil {
			return cuerrors.ErrorWithDetails(inst.Err, "failed to load CUE instance")
		}
	}
	return nil
}

// fillMetadata fills the CUE schema with the API version, kind, and metadata from the KRMInput configuration.
func fillMetadata(schema cue.Value, config *api.KRMInput) (cue.Value, error) {
	filledSchema := schema.FillPath(cue.ParsePath(APIVersionFillPath), config.APIVersion)
	filledSchema = filledSchema.FillPath(cue.ParsePath(KindFillPath), config.Kind)

	meta, err := api.IntoCueValue(schema.Context(), config.ObjectMeta)
	if err != nil {
		return cue.Value{}, cuerrors.ErrorWithDetails(err, "failed to convert ObjectMeta into CUE value")
	}

	filledSchema = filledSchema.FillPath(cue.ParsePath(MetadataFillPath), meta)
	return filledSchema, nil
}

// shouldActAsValidator checks if the KRMInput configuration has the validator annotation set.
func shouldActAsValidator(config *api.KRMInput) bool {
	return config.Annotations != nil &&
		config.Annotations[ValidatorAnnotationKey] == ValidatorAnnotationValue
}
