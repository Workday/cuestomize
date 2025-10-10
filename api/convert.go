package api

import (
	"context"
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/Workday/cuestomize/internal/pkg/cuerrors"
)

// CueConvertable is an interface that types can implement to convert themselves into a cue.Value.
type CueConvertable interface {
	IntoCueValue(ctx context.Context, cueCtx *cue.Context) (*cue.Value, error)
}

// IntoCueValue is a function that attempts to convert a given value into a cue.Value.
func IntoCueValue(ctx context.Context, cueCtx *cue.Context, v any) (*cue.Value, error) {
	detailer := cuerrors.FromContextOrDefault(ctx)
	asBytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	value := cueCtx.CompileBytes(asBytes)
	if value.Err() != nil {
		return nil, detailer.ErrorWithDetails(value.Err(), "failed to compile value into CUE")
	}
	return &value, nil
}
