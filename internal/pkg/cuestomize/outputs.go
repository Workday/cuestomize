package cuestomize

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"github.com/Workday/cuestomize/internal/pkg/cuerrors"
	"github.com/rs/zerolog/log"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// processOutputs processes the outputs from the CUE model and appends them to the output slice.
func processOutputs(unified cue.Value, items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	outputsValue := unified.LookupPath(cue.ParsePath(OutputsPath))
	if !outputsValue.Exists() {
		return nil, fmt.Errorf("'%s' not found in unified CUE instance", OutputsPath)
	} else if outputsValue.Err() != nil {
		return nil, cuerrors.ErrorWithDetails(outputsValue.Err(), "failed to lookup '%s' in unified CUE instance", OutputsPath)
	}
	outputsIter, err := getIter(outputsValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get iterator over '%s' in unified CUE instance: %v", OutputsPath, err)
	}

	for outputsIter.Next() {
		item := outputsIter.Value()

		rNode, err := cueValueToRNode(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert CUE value to kyaml.RNode: %w", err)
		}

		log.Debug().Str("apiVersion", rNode.GetApiVersion()).Str("kind", rNode.GetKind()).
			Str("namespace", rNode.GetNamespace()).Str("name", rNode.GetName()).Msg("adding item to output resources")
		items = append(items, rNode)
	}
	return items, nil
}

// getIter returns a cue.Iterator over a cue.Value of kind list or struct.
// It returns an error if the value is not a list nor a struct.
func getIter(value cue.Value) (*cue.Iterator, error) {
	kind := value.Kind()
	switch kind {
	case cue.ListKind:
		iter, _ := value.List()
		return &iter, nil
	case cue.StructKind:
		iter, _ := value.Fields()
		return iter, nil
	default:
		return nil, fmt.Errorf("value is not a list nor a struct, got: %s", kind)
	}
}

// cueValueToRNode converts a CUE value to a kyaml.RNode.
func cueValueToRNode(value *cue.Value) (*kyaml.RNode, error) {
	asBytes, err := yaml.Encode(*value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CUE value as YAML: %w", err)
	}

	rNode, err := kyaml.Parse(string(asBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse item as kyaml.RNode: %w", err)
	}

	return rNode, nil
}
