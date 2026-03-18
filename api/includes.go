package api

import (
	"cuelang.org/go/cue"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// Includes is a map that holds manifests, indexed by their API version, kind, namespace, and name respectively.
type Includes map[string]map[string]map[string]map[string]interface{}

// IntoCueValue tries to convert the Includes into a CUE value.
func (i Includes) IntoCueValue(cueCtx *cue.Context) (*cue.Value, error) {
	return IntoCueValue(cueCtx, i)
}

// Add adds an include to the Includes map.
func (i Includes) Add(include *kyaml.RNode) {
	i.initialiseMap(include)

	i[include.GetApiVersion()][include.GetKind()][include.GetNamespace()][include.GetName()] = any(include)
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
