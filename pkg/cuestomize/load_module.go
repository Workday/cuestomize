package cuestomize

import (
	"context"
	"fmt"

	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/load"
)

// LoadCUEModel loads a CUE model from the specified path and returns the instances.
func LoadCUEModel(ctx context.Context, path string) ([]*build.Instance, error) {
	cfg := &load.Config{Dir: path}
	instances := load.Instances([]string{"."}, cfg)
	if len(instances) == 0 {
		return nil, fmt.Errorf("no CUE instances found")
	}

	return instances, CheckInstances(ctx, instances)
}
