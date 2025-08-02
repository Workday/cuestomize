package api

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// Includes is a map that holds manifests, indexed by their API version, kind, namespace, and name respectively.
type Includes map[string]map[string]map[string]map[string]interface{}

// IntoCueValue tries to convert the Includes into a CUE value.
func (i Includes) IntoCueValue(ctx *cue.Context) (*cue.Value, error) {
	return IntoCueValue(ctx, i)
}

// Add adds an include to the Includes map.
func (i Includes) Add(include *kyaml.RNode) error {
	i.initialiseMap(include)
	obj, err := toMap(include)
	if err != nil {
		return fmt.Errorf("failed to convert item to map: %w", err)
	}
	i[include.GetApiVersion()][include.GetKind()][include.GetNamespace()][include.GetName()] = obj
	return nil
}

func (i Includes) initialiseMap(include *kyaml.RNode) {
	apiVersion := include.GetApiVersion()
	kind := include.GetKind()
	name := include.GetName()
	namespace := include.GetNamespace()

	// apiVersion
	if _, ok := i[apiVersion]; !ok {
		i[apiVersion] = make(map[string]map[string]map[string]interface{})
	}
	// kind
	if _, ok := i[apiVersion][kind]; !ok {
		i[apiVersion][kind] = make(map[string]map[string]interface{})
	}
	// namespace
	if _, ok := i[apiVersion][kind][namespace]; !ok {
		i[apiVersion][kind][namespace] = make(map[string]interface{})
	}
	// name
	if _, ok := i[apiVersion][kind][namespace][name]; !ok {
		i[apiVersion][kind][namespace][name] = make(map[string]interface{})
	}
}

// toMap converts a kyaml.RNode YAML contento to a map[string]interface{}.
func toMap(include *kyaml.RNode) (map[string]interface{}, error) {
	marshalled, err := include.MarshalJSON()
	if err != nil {
		return nil, err
	}
	obj := make(map[string]interface{})
	if err = json.Unmarshal(marshalled, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}
