package api

import (
	"oras.land/oras-go/v2/registry"
	"sigs.k8s.io/kustomize/api/types"
)

// RemoteModule defines the structure to describe a remote CUE module to fetch from an OCI registry.
type RemoteModule struct {
	// Ref is the full OCI reference in the format: registry/repo:tag
	// Example: ghcr.io/workday/my-module:v1.0.0
	// If Ref is specified, it takes precedence over Registry, Repo, and Tag.
	Ref string `yaml:"ref,omitempty" json:"ref,omitempty"`

	// Registry is the OCI registry hosting the module.
	//
	// Deprecated: Use Ref instead. Registry will be removed in a future version.
	Registry string `yaml:"registry,omitempty" json:"registry,omitempty"`
	// Repo is the repository path within the OCI registry.
	//
	// Deprecated: Use Ref instead. Repo will be removed in a future version.
	Repo string `yaml:"repo,omitempty" json:"repo,omitempty"`
	// Tag is the version tag of the module.
	//
	// Deprecated: Use Ref instead. Tag will be removed in a future version.
	Tag string `yaml:"tag,omitempty" json:"tag,omitempty"`

	Auth      *types.Selector `yaml:"auth,omitempty" json:"auth,omitempty"`
	PlainHTTP bool            `yaml:"plainHTTP,omitempty" json:"plainHTTP,omitempty"`
}

// GetRegistry returns the registry, preferring Ref over the deprecated Registry field.
func (r *RemoteModule) GetRegistry() (string, error) {
	if r.Ref != "" {
		ref, err := registry.ParseReference(r.Ref)
		return ref.Registry, err
	}
	return r.Registry, nil
}

// GetRepo returns the repository, preferring Ref over the deprecated Repo field.
func (r *RemoteModule) GetRepo() (string, error) {
	if r.Ref != "" {
		ref, err := registry.ParseReference(r.Ref)
		return ref.Repository, err
	}
	return r.Repo, nil
}

// GetTag returns the tag, preferring Ref over the deprecated Tag field.
func (r *RemoteModule) GetTag() (string, error) {
	if r.Ref != "" {
		ref, err := registry.ParseReference(r.Ref)
		return ref.Reference, err
	}
	return r.Tag, nil
}
