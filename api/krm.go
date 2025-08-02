package api

import (
	"fmt"

	"cuelang.org/go/cue"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// KRMInput is the input structure consumed by the KRM model.
type KRMInput struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Input contains the KRM input specification.
	Input    map[string]interface{} `yaml:"input" json:"input"`
	Includes []types.Selector       `yaml:"includes,omitempty" json:"includes,omitempty"`
}

// ExtractIncludes populates the includes structure from the provided KRMInput and items.
// It searches items for matches against the includes defined in the KRMInput's spec
// and returns the includes map.
func ExtractIncludes(krm *KRMInput, items []*kyaml.RNode) (Includes, error) {
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
			log.Warn().Msg(fmt.Sprintf("no items matched for selector: %s", sel.String()))
		}
	}

	return includes, nil
}

// IntoCueValue tries to convert the KRMInput into a CUE value.
// The method will not convert the whole KRMInput to a CUE value, but only the Input field.
// This is because the KRMInput part that needs to be passed to the CUE model is entirely
// contained in the Input field.
func (i *KRMInput) IntoCueValue(ctx *cue.Context) (*cue.Value, error) {
	return IntoCueValue(ctx, i.Input)
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
