package api

import (
	"testing"

	"github.com/Workday/cuestomize/internal/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	TestKRMInputFileName = "krm-input.yaml"
	TestItemsFileName    = "items.yaml"
)

func TestKRMInput_ExtractIncludes(t *testing.T) {
	tests := []struct {
		name           string
		testdataDir    string
		expectedError  bool
		errorSubstring string
	}{
		{
			name:          "all includes found",
			testdataDir:   "../testdata/api/krm/ok-includes",
			expectedError: false,
		},
		{
			name:          "ok no matching includes",
			testdataDir:   "../testdata/api/krm/ok-no-matching-includes",
			expectedError: false,
		},
		{
			name:          "no includes",
			testdataDir:   "../testdata/api/krm/ok-no-includes",
			expectedError: false,
		},
		{
			name:           "malformed selector",
			testdataDir:    "../testdata/api/krm/nok-malformed-selector",
			expectedError:  true,
			errorSubstring: "failed to match item against selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			krmInput := testhelpers.LoadFromFile[KRMInput](t, tt.testdataDir+"/"+TestKRMInputFileName)
			items := testhelpers.LoadResourceList(t, tt.testdataDir+"/"+TestKRMInputFileName, tt.testdataDir+"/"+TestItemsFileName)

			includes, err := ExtractIncludes(t.Context(), krmInput, items)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorSubstring != "" {
					require.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, includes, "includes should not be nil")

				// TODO: verify the contents of the KRMInput.Spec after conversion
			}
		})
	}
}

func TestFindAuthSecret(t *testing.T) {
	tests := []struct {
		name          string
		selector      types.Selector
		items         []*kyaml.RNode
		expectedName  string
		expectedError string
	}{
		{
			name: "selector kind is not Secret",
			selector: types.Selector{
				ResId: resid.ResId{
					Gvk: resid.Gvk{Kind: "ConfigMap"},
				},
			},
			items:         []*kyaml.RNode{},
			expectedError: `kind must be Secret, got: "ConfigMap"`,
		},
		{
			name: "secret not found",
			selector: types.Selector{
				ResId: resid.ResId{
					Gvk: resid.Gvk{Kind: "Secret"}, Name: "my-secret",
				},
			},
			items: []*kyaml.RNode{
				createTestNode(t, "v1", "Secret", "default", "other-secret"),
			},
			expectedError: "no items matched for selector",
		},
		{
			name: "secret found by name",
			selector: types.Selector{
				ResId: resid.ResId{
					Gvk: resid.Gvk{Kind: "Secret"}, Name: "my-secret",
				},
			},
			items: []*kyaml.RNode{
				createTestNode(t, "v1", "Secret", "default", "other-secret"),
				createTestNode(t, "v1", "Secret", "default", "my-secret"),
			},
			expectedName: "my-secret",
		},
		{
			name: "secret found by name and namespace",
			selector: types.Selector{
				ResId: resid.ResId{
					Gvk:       resid.Gvk{Kind: "Secret"},
					Name:      "my-secret",
					Namespace: "my-ns",
				},
			},
			items: []*kyaml.RNode{
				createTestNode(t, "v1", "Secret", "default", "my-secret"),
				createTestNode(t, "v1", "Secret", "my-ns", "my-secret"),
			},
			expectedName: "my-secret",
		},
		{
			name: "secret found by label",
			selector: types.Selector{
				ResId:         resid.ResId{Gvk: resid.Gvk{Kind: "Secret"}},
				LabelSelector: "app=my-app",
			},
			items: []*kyaml.RNode{
				func() *kyaml.RNode {
					t.Helper()
					node := createTestNode(t, "v1", "Secret", "default", "labeled-secret")
					require.NoError(t, node.SetLabels(map[string]string{"app": "my-app"}))
					return node
				}(),
			},
			expectedName: "labeled-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := findAuthSecret(&tt.selector, tt.items)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, secret)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, tt.expectedName, secret.Name)
				if tt.selector.Namespace != "" {
					assert.Equal(t, tt.selector.Namespace, secret.Namespace)
				}
			}
		})
	}
}
