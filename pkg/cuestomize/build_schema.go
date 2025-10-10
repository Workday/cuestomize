// Package cuestomize provides the Cuestomize functionality.
package cuestomize

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"github.com/Workday/cuestomize/internal/pkg/cuerrors"
)

// BuildCUEModelSchema builds a CUE model from the provided instances and returns the unified schema.
func BuildCUEModelSchema(ctx context.Context, cueCtx *cue.Context, instances []*build.Instance) (*cue.Value, error) {
	detailer := cuerrors.FromContextOrDefault(ctx)

	values, err := cueCtx.BuildInstances(instances)
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
			return nil, detailer.ErrorWithDetails(schema.Err(), "failed to unify CUE model with [%v]", instances[i].BuildFiles)
		}
	}

	return &schema, nil
}
