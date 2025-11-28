package cuestomize

import (
	"context"

	"cuelang.org/go/cue/build"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/cuerrors"
)

const (
	// ValidatorAnnotationKey is the annotation key that marks a CUE function as a validator.
	ValidatorAnnotationKey = "config.cuestomize.io/validator"
	// ValidatorAnnotationValue is the value of the annotation that marks a CUE function as a validator.
	ValidatorAnnotationValue = "true"
)

// CheckInstances checks if any of the instances have an error and returns an error if so.
func CheckInstances(ctx context.Context, instances []*build.Instance) error {
	detailer := cuerrors.FromContextOrEmpty(ctx)

	for _, inst := range instances {
		if inst.Err != nil {
			return detailer.ErrorWithDetails(inst.Err, "failed to load CUE instance")
		}
	}
	return nil
}

// ShouldActAsValidator checks if the KRMInput configuration has the validator annotation set.
func ShouldActAsValidator(config *api.KRMInput) bool {
	return config.Annotations != nil &&
		config.Annotations[ValidatorAnnotationKey] == ValidatorAnnotationValue
}
