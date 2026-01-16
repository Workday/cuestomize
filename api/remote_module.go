package api

import (
	"fmt"
	"strings"

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
		registry, _, _, err := parseOCIRef(r.Ref)
		return registry, err
	}
	return r.Registry, nil
}

// GetRepo returns the repository, preferring Ref over the deprecated Repo field.
func (r *RemoteModule) GetRepo() (string, error) {
	if r.Ref != "" {
		_, repo, _, err := parseOCIRef(r.Ref)
		return repo, err
	}
	return r.Repo, nil
}

// GetTag returns the tag, preferring Ref over the deprecated Tag field.
func (r *RemoteModule) GetTag() (string, error) {
	if r.Ref != "" {
		_, _, tag, err := parseOCIRef(r.Ref)
		return tag, err
	}
	return r.Tag, nil
}

// parseOCIRef parses a full OCI reference into its components.
// Example: "ghcr.io/workday/my-module:v1.0.0" -> ("ghcr.io", "workday/my-module", "v1.0.0", nil)
// If no tag is specified, returns "latest" as the tag.
func parseOCIRef(ref string) (registry, repo, tag string, err error) {
	// Split by colon to separate tag
	tagParts := strings.Split(ref, ":")
	if len(tagParts) > 2 {
		return "", "", "", fmt.Errorf("invalid OCI reference format: %s (multiple colons found)", ref)
	}

	// Get tag or default to "latest"
	if len(tagParts) == 2 {
		tag = tagParts[1]
	} else {
		tag = "latest"
	}

	// Split by slash to separate registry and repo
	pathWithRegistry := tagParts[0]
	parts := strings.Split(pathWithRegistry, "/")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid OCI reference format: %s (expected format: registry/repo:tag)", ref)
	}

	registry = parts[0]
	repo = strings.Join(parts[1:], "/")

	return registry, repo, tag, nil
}
