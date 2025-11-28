package cuestomize

import (
	"context"

	"cuelang.org/go/cue"
	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/cuerrors"
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

// FillMetadata fills the CUE schema with the API version, kind, and metadata from the KRMInput configuration.
func FillMetadata(ctx context.Context, schema cue.Value, config *api.KRMInput) (cue.Value, error) {
	detailer := cuerrors.FromContextOrEmpty(ctx)

	filledSchema := schema.FillPath(cue.ParsePath(APIVersionFillPath), config.APIVersion)
	filledSchema = filledSchema.FillPath(cue.ParsePath(KindFillPath), config.Kind)

	meta, err := api.IntoCueValue(schema.Context(), config.ObjectMeta)
	if err != nil {
		return cue.Value{}, detailer.ErrorWithDetails(err, "failed to convert ObjectMeta into CUE value")
	}

	filledSchema = filledSchema.FillPath(cue.ParsePath(MetadataFillPath), meta)
	return filledSchema, nil
}
