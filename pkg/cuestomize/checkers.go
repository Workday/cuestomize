package cuestomize

import (
	"cuelang.org/go/cue/build"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/internal/pkg/cuerrors"
)

const (
	// ValidatorAnnotationKey is the annotation key that marks a CUE function as a validator.
	ValidatorAnnotationKey = "config.cuestomize.io/validator"
	// ValidatorAnnotationValue is the value of the annotation that marks a CUE function as a validator.
	ValidatorAnnotationValue = "true"
)

// CheckInstances checks if any of the instances have an error and returns an error if so.
func CheckInstances(instances []*build.Instance) error {
	for _, inst := range instances {
		if inst.Err != nil {
			return cuerrors.ErrorWithDetails(inst.Err, "failed to load CUE instance")
		}
	}
	return nil
}

// ShouldActAsValidator checks if the KRMInput configuration has the validator annotation set.
func ShouldActAsValidator(config *api.KRMInput) bool {
	return config.Annotations != nil &&
		config.Annotations[ValidatorAnnotationKey] == ValidatorAnnotationValue
}
