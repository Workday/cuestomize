package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestIncludes_Add(t *testing.T) {
	tests := []struct {
		name      string
		nodes     []*kyaml.RNode
		expectErr bool
		validate  func(t *testing.T, includes Includes)
	}{
		{
			name: "add single item",
			nodes: []*kyaml.RNode{
				createTestNode(t, "v1", "ConfigMap", "default", "config1"),
			},
			expectErr: false,
			validate: func(t *testing.T, includes Includes) {
				assert.Len(t, includes, 1)
				assert.Contains(t, includes, "v1")
				assert.Contains(t, includes["v1"], "ConfigMap")
				assert.Contains(t, includes["v1"]["ConfigMap"], "default")
				assert.Contains(t, includes["v1"]["ConfigMap"]["default"], "config1")
				assert.NotNil(t, includes["v1"]["ConfigMap"]["default"]["config1"])
			},
		},
		{
			name: "add multiple items with different apiVersions",
			nodes: []*kyaml.RNode{
				createTestNode(t, "v1", "ConfigMap", "default", "config1"),
				createTestNode(t, "apps/v1", "Deployment", "app", "deploy1"),
			},
			expectErr: false,
			validate: func(t *testing.T, includes Includes) {
				assert.Len(t, includes, 2) // v1 and apps/v1

				// Check v1/ConfigMap/default/config1
				assert.Contains(t, includes, "v1")
				assert.Contains(t, includes["v1"], "ConfigMap")
				assert.Contains(t, includes["v1"]["ConfigMap"]["default"], "config1")

				// Check apps/v1/Deployment/app/deploy1
				assert.Contains(t, includes, "apps/v1")
				assert.Contains(t, includes["apps/v1"], "Deployment")
				assert.Contains(t, includes["apps/v1"]["Deployment"]["app"], "deploy1")
			},
		},
		{
			name: "add multiple items with different properties",
			nodes: []*kyaml.RNode{
				createTestNode(t, "v1", "ConfigMap", "default", "config1"),
				createTestNode(t, "v1", "Secret", "default", "secret1"),
				createTestNode(t, "apps/v1", "Deployment", "app", "deploy1"),
			},
			expectErr: false,
			validate: func(t *testing.T, includes Includes) {
				assert.Len(t, includes, 2) // v1 and apps/v1

				// Check v1/ConfigMap/default/config1
				assert.Contains(t, includes, "v1")
				assert.Contains(t, includes["v1"], "ConfigMap")
				assert.Contains(t, includes["v1"]["ConfigMap"]["default"], "config1")

				// Check v1/Secret/default/secret1
				assert.Contains(t, includes["v1"], "Secret")
				assert.Contains(t, includes["v1"]["Secret"]["default"], "secret1")

				// Check apps/v1/Deployment/app/deploy1
				assert.Contains(t, includes, "apps/v1")
				assert.Contains(t, includes["apps/v1"], "Deployment")
				assert.Contains(t, includes["apps/v1"]["Deployment"]["app"], "deploy1")
			},
		},
		{
			name: "add items with same apiVersion and kind but different namespace",
			nodes: []*kyaml.RNode{
				createTestNode(t, "v1", "ConfigMap", "default", "config1"),
				createTestNode(t, "v1", "ConfigMap", "kube-system", "config2"),
			},
			expectErr: false,
			validate: func(t *testing.T, includes Includes) {
				assert.Len(t, includes["v1"]["ConfigMap"], 2)
				assert.Contains(t, includes["v1"]["ConfigMap"]["default"], "config1")
				assert.Contains(t, includes["v1"]["ConfigMap"]["kube-system"], "config2")
			},
		},
		{
			name: "add items with same apiVersion, kind, namespace but different names",
			nodes: []*kyaml.RNode{
				createTestNode(t, "v1", "ConfigMap", "default", "config1"),
				createTestNode(t, "v1", "ConfigMap", "default", "config2"),
			},
			expectErr: false,
			validate: func(t *testing.T, includes Includes) {
				assert.Len(t, includes["v1"]["ConfigMap"]["default"], 2)
				assert.Contains(t, includes["v1"]["ConfigMap"]["default"], "config1")
				assert.Contains(t, includes["v1"]["ConfigMap"]["default"], "config2")
			},
		},
		{
			name: "add items in reverse order",
			nodes: []*kyaml.RNode{
				createTestNode(t, "apps/v1", "Deployment", "app", "deploy1"),
				createTestNode(t, "v1", "Secret", "default", "secret1"),
				createTestNode(t, "v1", "ConfigMap", "default", "config1"),
			},
			expectErr: false,
			validate: func(t *testing.T, includes Includes) {
				assert.Len(t, includes, 2) // v1 and apps/v1

				// Verify all items exist
				assert.Contains(t, includes["v1"]["ConfigMap"]["default"], "config1")
				assert.Contains(t, includes["v1"]["Secret"]["default"], "secret1")
				assert.Contains(t, includes["apps/v1"]["Deployment"]["app"], "deploy1")
			},
		},
		{
			name: "add items with empty namespace",
			nodes: []*kyaml.RNode{
				createTestNode(t, "v1", "ConfigMap", "", "config1"),
			},
			expectErr: false,
			validate: func(t *testing.T, includes Includes) {
				assert.Contains(t, includes["v1"]["ConfigMap"][""], "config1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			includes := make(Includes)

			for _, node := range tt.nodes {
				err := includes.Add(node)
				if tt.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}

			if tt.validate != nil {
				tt.validate(t, includes)
			}
		})
	}
}

func createTestNode(t *testing.T, apiVersion, kind, namespace, name string) *kyaml.RNode {
	t.Helper()

	yamlObj := map[string]interface{}{}
	yamlObj["apiVersion"] = apiVersion
	yamlObj["kind"] = kind
	yamlObj["metadata"] = map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	}

	node, err := kyaml.FromMap(yamlObj)
	require.NoError(t, err)
	return node
}
