package api

import (
	"context"
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	registryauth "github.com/Workday/cuestomize/internal/pkg/registry_auth"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"oras.land/oras-go/v2/registry/remote/auth"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// KRMInput is the input structure consumed by the KRM model.
type KRMInput struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Input contains the KRM input specification.
	Input        map[string]interface{} `yaml:"input" json:"input"`
	Includes     []types.Selector       `yaml:"includes,omitempty" json:"includes,omitempty"`
	RemoteModule *RemoteModule          `yaml:"remoteModule,omitempty" json:"remoteModule,omitempty"`
}

// ExtractIncludes populates the includes structure from the provided KRMInput and items.
// It searches items for matches against the includes defined in the KRMInput's spec
// and returns the includes map.
func ExtractIncludes(ctx context.Context, krm *KRMInput, items []*kyaml.RNode) (Includes, error) {
	log := logr.FromContextOrDiscard(ctx)

	includes := make(Includes)

	for _, sel := range krm.Includes {
		includesCount := 0
		for _, item := range items {
			itemMatches, err := ItemMatchReference(item, &sel)
			if err != nil {
				return nil, fmt.Errorf("failed to match item against selector [%v]: %w", sel.String(), err)
			}
			if itemMatches {
				includesCount++
				if err := includes.Add(item); err != nil {
					return nil, fmt.Errorf("failed to add include: %w", err)
				}
			}
		}
		if includesCount == 0 {
			log.V(-1).Info("no items matched for include selector", "selector", sel.String())
		}
	}

	return includes, nil
}

// IntoCueValue tries to convert the KRMInput into a CUE value.
// The method will not convert the whole KRMInput to a CUE value, but only the Input field.
// This is because the KRMInput part that needs to be passed to the CUE model is entirely
// contained in the Input field.
func (i *KRMInput) IntoCueValue(ctx context.Context, cueCtx *cue.Context) (*cue.Value, error) {
	return IntoCueValue(ctx, cueCtx, i.Input)
}

// GetRemoteClient returns a remote client based on the remote module configuration.
// If no authentication configuration is found, it returns nil. A nil client is a valid value,
// check the error return value for actual errors.
func (i *KRMInput) GetRemoteClient(items []*kyaml.RNode) (*auth.Client, error) {
	var secret *corev1.Secret
	var err error
	if i.RemoteModule.Auth != nil {
		secret, err = findAuthSecret(i.RemoteModule.Auth, items)
		if err != nil {
			return nil, fmt.Errorf("failed to find auth secret: %w", err)
		}
	}

	return registryauth.ConfigureClient(i.RemoteModule.Registry, secret)
}

// ItemMatchReference checks if the given item matches the provided selector.
func ItemMatchReference(item *kyaml.RNode, sel *types.Selector) (bool, error) {
	matchesLabel, err := item.MatchesLabelSelector(sel.LabelSelector)
	if err != nil {
		return false, fmt.Errorf("failed to match label selector: %w", err)
	}

	matchesAnnotation, err := item.MatchesAnnotationSelector(sel.AnnotationSelector)
	if err != nil {
		return false, fmt.Errorf("failed to match annotation selector: %w", err)
	}

	if !matchesLabel || !matchesAnnotation {
		return false, nil
	}

	selRe, err := types.NewSelectorRegex(sel)
	if err != nil {
		return false, fmt.Errorf("failed to create selector regex: %w", err)
	}
	return selRe.MatchGvk(resid.GvkFromNode(item)) &&
		selRe.MatchName(item.GetName()) &&
		selRe.MatchNamespace(item.GetNamespace()), nil
}

// findAuthSecret searches items for a Secret that matches the provided selector.
// It returns the found Secret or an error if no match is found.
func findAuthSecret(sel *types.Selector, items []*kyaml.RNode) (*corev1.Secret, error) {
	if sel.Kind != "Secret" {
		return nil, fmt.Errorf(`kind must be Secret, got: "%s"`, sel.Kind)
	}

	for _, item := range items {
		matches, err := ItemMatchReference(item, sel)
		if err != nil {
			return nil, fmt.Errorf("failed to match item against selector: %w", err)
		}
		if matches {
			// convert to corev1.Secret
			bytes, err := item.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal item to JSON: %w", err)
			}

			secret := &corev1.Secret{}
			if err := json.Unmarshal(bytes, secret); err != nil {
				return nil, fmt.Errorf("failed to unmarshal item to corev1.Secret: %w", err)
			}
			return secret, nil
		}
	}

	return nil, fmt.Errorf("no items matched for selector [%s]", sel.String())
}
