// Package api provides utilities for converting Go values into CUE values and defines
// interfaces for types that can be converted to CUE format.
package api

import (
	"cuelang.org/go/cue"
)

// IntoCueValue is a function that attempts to convert a given value into a cue.Value.
func IntoCueValue(cueCtx *cue.Context, v any) (*cue.Value, error) {
	value := cueCtx.Encode(v)
	return &value, value.Err()
}
