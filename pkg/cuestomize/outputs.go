package cuestomize

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"github.com/Workday/cuestomize/pkg/cuerrors"

	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type OutputOptions struct {
	// AllowEdit allows output resources to replace existing items in the stream if a matching ResId is found.
	AllowEdit bool
}

// ProcessOutputs processes the outputs from the CUE model and appends them to the output slice.
// When allowEdit is true, output resources that match an existing item by ResId (GVK + namespace + name)
// will replace the existing item in place; otherwise they are appended.
func ProcessOutputs(ctx context.Context, unified cue.Value, items []*kyaml.RNode, opts OutputOptions) ([]*kyaml.RNode, error) {
	detailer := cuerrors.FromContextOrEmpty(ctx)

	outputsValue := unified.LookupPath(cue.ParsePath(OutputsPath))
	if !outputsValue.Exists() {
		return nil, fmt.Errorf("'%s' not found in unified CUE instance", OutputsPath)
	} else if outputsValue.Err() != nil {
		return nil, detailer.ErrorWithDetails(outputsValue.Err(), "failed to lookup '%s' in unified CUE instance", OutputsPath)
	}
	outputsIter, err := getIter(outputsValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get iterator over '%s' in unified CUE instance: %v", OutputsPath, err)
	}

	if !opts.AllowEdit {
		return appendOutputs(ctx, outputsIter, items)
	}
	return editOutputs(ctx, outputsIter, items)
}

// appendOutputs appends all CUE outputs to the items slice (generate-only mode).
func appendOutputs(ctx context.Context, outputsIter *cue.Iterator, items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	for outputsIter.Next() {
		item := outputsIter.Value()

		rNode, err := cueValueToRNode(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert CUE value to kyaml.RNode: %w", err)
		}

		items = append(items, rNode)
	}
	return items, nil
}

// editOutputs replaces existing resources in the items stream if a matching
// resource (by ResId) is found, or appends new ones.
func editOutputs(ctx context.Context, outputsIter *cue.Iterator, items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	residMap := make(map[resid.ResId]*kyaml.RNode)

	for outputsIter.Next() {
		item := outputsIter.Value()

		rNode, err := cueValueToRNode(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert CUE value to kyaml.RNode: %w", err)
		}

		rid := residFromRNode(rNode)

		if _, found := residMap[rid]; found {
			return nil, fmt.Errorf("duplicate output resource with ResId '%s' found in CUE model", rid)
		}
	}

	for i := range items {
		itemRid := residFromRNode(items[i])
		if cueOutputRNode, found := residMap[itemRid]; found {
			items[i] = cueOutputRNode
			delete(residMap, itemRid)
		}
	}

	// append any remaining CUE output resource
	for _, rNode := range residMap {
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

func residFromRNode(rNode *kyaml.RNode) resid.ResId {
	return resid.NewResIdWithNamespace(resid.GvkFromNode(rNode), rNode.GetName(), rNode.GetNamespace())
}
