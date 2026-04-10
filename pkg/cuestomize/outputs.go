package cuestomize

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"github.com/Workday/cuestomize/pkg/cuerrors"
	"github.com/go-logr/logr"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// ProcessOutputs processes the outputs from the CUE model and appends them to the output slice.
// When allowEdit is true, output resources that match an existing item by ResId (GVK + namespace + name)
// will replace the existing item in place; otherwise they are appended.
func ProcessOutputs(ctx context.Context, unified cue.Value, items []*kyaml.RNode, allowEdit bool) ([]*kyaml.RNode, error) {
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

	if !allowEdit {
		return appendOutputs(ctx, outputsIter, items)
	}
	return editOutputs(ctx, outputsIter, items)
}

// appendOutputs appends all CUE outputs to the items slice (generate-only mode).
func appendOutputs(ctx context.Context, outputsIter *cue.Iterator, items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	log := logr.FromContextOrDiscard(ctx)

	for outputsIter.Next() {
		item := outputsIter.Value()

		rNode, err := cueValueToRNode(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert CUE value to kyaml.RNode: %w", err)
		}

		log.V(4).Info("adding item to output resources",
			"kind", rNode.GetKind(), "apiVersion", rNode.GetApiVersion(), "namespace", rNode.GetNamespace(), "name", rNode.GetName())
		items = append(items, rNode)
	}
	return items, nil
}

// editOutputs replaces existing resources in the items stream if a matching
// resource (by ResId) is found, or appends new ones.
func editOutputs(ctx context.Context, outputsIter *cue.Iterator, items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	log := logr.FromContextOrDiscard(ctx)

	rf := resource.NewFactory(nil)
	rf.IncludeLocalConfigs = true

	rmf := resmap.NewFactory(rf)

	streamRM, err := rmf.NewResMapFromRNodeSlice(items)
	if err != nil {
		return nil, fmt.Errorf("failed to create ResMap from input items: %w", err)
	}

	for outputsIter.Next() {
		item := outputsIter.Value()

		rNode, err := cueValueToRNode(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert CUE value to kyaml.RNode: %w", err)
		}

		rid := resid.NewResIdWithNamespace(resid.GvkFromNode(rNode), rNode.GetName(), rNode.GetNamespace())
		for i, _ := range items {
			itemRid := resid.NewResIdWithNamespace(resid.GvkFromNode(items[i]), items[i].GetName(), items[i].GetNamespace())
			if rid.Equals(itemRid) {
				items[i] = rNode
				break
			}
		}

		ress, err := rf.ResourcesFromRNodes([]*kyaml.RNode{rNode})
		if err != nil {
			return nil, fmt.Errorf("failed to convert RNode to Resource: %w", err)
		}
		res := ress[0]

		idx, err := streamRM.GetIndexOfCurrentId(res.CurId())
		if err != nil {
			return nil, fmt.Errorf("failed to look up resource %s in stream: %w", res.CurId(), err)
		}

		if idx >= 0 {
			log.V(4).Info("replacing item in output resources",
				"kind", rNode.GetKind(), "apiVersion", rNode.GetApiVersion(), "namespace", rNode.GetNamespace(), "name", rNode.GetName())
			if _, err := streamRM.Replace(res); err != nil {
				return nil, fmt.Errorf("failed to replace resource %s in stream: %w", res.CurId(), err)
			}
		} else {
			log.V(4).Info("adding item to output resources",
				"kind", rNode.GetKind(), "apiVersion", rNode.GetApiVersion(), "namespace", rNode.GetNamespace(), "name", rNode.GetName())
			if err := streamRM.Append(res); err != nil {
				return nil, fmt.Errorf("failed to append resource %s to stream: %w", res.CurId(), err)
			}
		}
	}

	return streamRM.ToRNodeSlice(), nil
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
