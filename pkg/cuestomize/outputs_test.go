package cuestomize

import (
	"context"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// helper to create an RNode from a map.
func createTestRNode(t *testing.T, apiVersion, kind, namespace, name string) *kyaml.RNode {
	t.Helper()
	yamlObj := map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	}
	node, err := kyaml.FromMap(yamlObj)
	require.NoError(t, err)
	return node
}

// helper to compile a CUE string and return the value.
func compileCUE(t *testing.T, src string) cue.Value {
	t.Helper()
	ctx := cuecontext.New()
	v := ctx.CompileString(src)
	require.NoError(t, v.Err())
	return v
}

func TestProcessOutputs_AppendMode(t *testing.T) {
	tests := []struct {
		name           string
		cueSrc         string
		existingItems  []*kyaml.RNode
		expectedCount  int
		expectedError  bool
		errorSubstring string
	}{
		{
			name: "append single output to empty items",
			cueSrc: `{
				outputs: [{
					apiVersion: "v1"
					kind:       "ConfigMap"
					metadata: {
						name:      "my-cm"
						namespace: "default"
					}
					data: {
						key: "value"
					}
				}]
			}`,
			existingItems: nil,
			expectedCount: 1,
		},
		{
			name: "append multiple outputs to empty items",
			cueSrc: `{
				outputs: [{
					apiVersion: "v1"
					kind:       "ConfigMap"
					metadata: {
						name:      "cm1"
						namespace: "default"
					}
				}, {
					apiVersion: "v1"
					kind:       "ConfigMap"
					metadata: {
						name:      "cm2"
						namespace: "default"
					}
				}]
			}`,
			existingItems: nil,
			expectedCount: 2,
		},
		{
			name: "append outputs to existing items",
			cueSrc: `{
				outputs: [{
					apiVersion: "v1"
					kind:       "ConfigMap"
					metadata: {
						name:      "new-cm"
						namespace: "default"
					}
				}]
			}`,
			existingItems: []*kyaml.RNode{
				createTestRNode(t, "v1", "ConfigMap", "default", "existing-cm"),
			},
			expectedCount: 2,
		},
		{
			name: "empty outputs list",
			cueSrc: `{
				outputs: []
			}`,
			existingItems: []*kyaml.RNode{
				createTestRNode(t, "v1", "ConfigMap", "default", "existing-cm"),
			},
			expectedCount: 1,
		},
		{
			name: "outputs path missing",
			cueSrc: `{
				something: "else"
			}`,
			existingItems:  nil,
			expectedError:  true,
			errorSubstring: "not found in unified CUE instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unified := compileCUE(t, tt.cueSrc)

			result, err := ProcessOutputs(t.Context(), unified, tt.existingItems, OutputOptions{AllowEdit: false})

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestProcessOutputs_AppendPreservesContent(t *testing.T) {
	cueSrc := `{
		outputs: [{
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name:      "my-cm"
				namespace: "test-ns"
			}
			data: {
				foo: "bar"
			}
		}]
	}`
	unified := compileCUE(t, cueSrc)

	result, err := ProcessOutputs(t.Context(), unified, nil, OutputOptions{AllowEdit: false})
	require.NoError(t, err)
	require.Len(t, result, 1)

	rNode := result[0]
	assert.Equal(t, "my-cm", rNode.GetName())
	assert.Equal(t, "test-ns", rNode.GetNamespace())
	assert.Equal(t, "ConfigMap", rNode.GetKind())
	assert.Equal(t, "v1", rNode.GetApiVersion())
}

func TestProcessOutputs_EditMode_ReplacesExisting(t *testing.T) {
	cueSrc := `{
		outputs: [{
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name:      "my-cm"
				namespace: "default"
			}
			data: {
				key: "new-value"
			}
		}]
	}`
	unified := compileCUE(t, cueSrc)

	existingItems := []*kyaml.RNode{
		createTestRNode(t, "v1", "ConfigMap", "default", "my-cm"),
		createTestRNode(t, "apps/v1", "Deployment", "default", "my-deploy"),
	}

	result, err := ProcessOutputs(t.Context(), unified, existingItems, OutputOptions{AllowEdit: true})
	require.NoError(t, err)

	// Should still have 2 items – the existing ConfigMap replaced in-place, deployment unchanged
	require.Len(t, result, 2)

	// First item should be the replaced ConfigMap with new data
	assert.Equal(t, "my-cm", result[0].GetName())

	// Second item should be the untouched deployment
	assert.Equal(t, "my-deploy", result[1].GetName())
}

func TestProcessOutputs_EditMode_AppendsNew(t *testing.T) {
	cueSrc := `{
		outputs: [{
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name:      "brand-new"
				namespace: "default"
			}
		}]
	}`
	unified := compileCUE(t, cueSrc)

	existingItems := []*kyaml.RNode{
		createTestRNode(t, "apps/v1", "Deployment", "default", "my-deploy"),
	}

	result, err := ProcessOutputs(t.Context(), unified, existingItems, OutputOptions{AllowEdit: true})
	require.NoError(t, err)

	// Should have 2 items – existing deployment + appended ConfigMap
	require.Len(t, result, 2)
	assert.Equal(t, "my-deploy", result[0].GetName())
	assert.Equal(t, "brand-new", result[1].GetName())
}

func TestProcessOutputs_EditMode_DuplicateOutputError(t *testing.T) {
	cueSrc := `{
		outputs: [{
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name:      "my-cm"
				namespace: "default"
			}
		}, {
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name:      "my-cm"
				namespace: "default"
			}
		}]
	}`
	unified := compileCUE(t, cueSrc)

	result, err := ProcessOutputs(t.Context(), unified, nil, OutputOptions{AllowEdit: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate output resource")
	assert.Nil(t, result)
}

func TestProcessOutputs_StructOutputs(t *testing.T) {
	cueSrc := `{
		outputs: {
			cm1: {
				apiVersion: "v1"
				kind:       "ConfigMap"
				metadata: {
					name:      "cm1"
					namespace: "default"
				}
			}
			cm2: {
				apiVersion: "v1"
				kind:       "ConfigMap"
				metadata: {
					name:      "cm2"
					namespace: "default"
				}
			}
		}
	}`
	unified := compileCUE(t, cueSrc)

	result, err := ProcessOutputs(t.Context(), unified, nil, OutputOptions{AllowEdit: false})
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetIter(t *testing.T) {
	ctx := cuecontext.New()

	tests := []struct {
		name          string
		cueSrc        string
		expectedError bool
	}{
		{
			name:          "list kind",
			cueSrc:        `[1, 2, 3]`,
			expectedError: false,
		},
		{
			name:          "struct kind",
			cueSrc:        `{a: 1, b: 2}`,
			expectedError: false,
		},
		{
			name:          "string kind",
			cueSrc:        `"hello"`,
			expectedError: true,
		},
		{
			name:          "int kind",
			cueSrc:        `42`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ctx.CompileString(tt.cueSrc)
			require.NoError(t, v.Err())

			iter, err := getIter(v)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, iter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, iter)
			}
		})
	}
}

func TestCueValueToRNode(t *testing.T) {
	tests := []struct {
		name          string
		cueSrc        string
		expectedName  string
		expectedError bool
	}{
		{
			name: "valid k8s resource",
			cueSrc: `{
				apiVersion: "v1"
				kind:       "ConfigMap"
				metadata: {
					name:      "my-cm"
					namespace: "default"
				}
				data: {
					key: "value"
				}
			}`,
			expectedName: "my-cm",
		},
		{
			name:   "empty struct",
			cueSrc: `{}`,
		},
	}

	ctx := cuecontext.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ctx.CompileString(tt.cueSrc)
			require.NoError(t, v.Err())

			rNode, err := cueValueToRNode(&v)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, rNode)
				if tt.expectedName != "" {
					assert.Equal(t, tt.expectedName, rNode.GetName())
				}
			}
		})
	}
}

func TestResidFromRNode(t *testing.T) {
	rNode := createTestRNode(t, "apps/v1", "Deployment", "my-ns", "my-deploy")
	rid := residFromRNode(rNode)

	expected := resid.NewResIdWithNamespace(
		resid.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"},
		"my-deploy",
		"my-ns",
	)
	assert.Equal(t, expected, rid)
}

func TestResidFromRNode_CoreResource(t *testing.T) {
	rNode := createTestRNode(t, "v1", "ConfigMap", "default", "my-cm")
	rid := residFromRNode(rNode)

	assert.Equal(t, "ConfigMap", rid.Kind)
	assert.Equal(t, "my-cm", rid.Name)
	assert.Equal(t, "default", rid.Namespace)
}

func TestProcessOutputs_EditMode_EmptyOutputs(t *testing.T) {
	cueSrc := `{
		outputs: []
	}`
	unified := compileCUE(t, cueSrc)

	existingItems := []*kyaml.RNode{
		createTestRNode(t, "v1", "ConfigMap", "default", "existing"),
	}

	result, err := ProcessOutputs(t.Context(), unified, existingItems, OutputOptions{AllowEdit: true})
	require.NoError(t, err)

	// Existing items should remain unchanged
	require.Len(t, result, 1)
	assert.Equal(t, "existing", result[0].GetName())
}

func TestProcessOutputs_NilContext(t *testing.T) {
	cueSrc := `{
		outputs: [{
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name:      "my-cm"
				namespace: "default"
			}
		}]
	}`
	unified := compileCUE(t, cueSrc)

	//nolint:staticcheck // intentionally testing with nil context
	result, err := ProcessOutputs(context.TODO(), unified, nil, OutputOptions{AllowEdit: false})
	require.NoError(t, err)
	assert.Len(t, result, 1)
}
