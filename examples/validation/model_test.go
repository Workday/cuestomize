package validation

import (
	"os"
	"testing"

	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/cuestomize"
	"github.com/Workday/cuestomize/pkg/cuestomize/model"
	"github.com/stretchr/testify/assert"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	yaml "sigs.k8s.io/yaml"
)

var (
	wrongConfigString = `apiVersion: cuestomize.dev/v1alpha1
kind: Validator
metadata:
  name: example
  annotations:
    config.cuestomize.io/validator: "true"
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/workday/cuestomize:latest
        network: true
input:
  unknownField: true
`
)

// TestValidationModel tests the validation model with various configurations and resources.
func TestValidationModel(t *testing.T) {
	okConfig := loadConfigFromFile(t, "./kustomize/validator.yaml")
	invalidConfig := loadConfig(t, wrongConfigString)
	okDeploy := loadResource(t, "./kustomize/deployment.yaml")
	invalidDeploy := loadResource(t, "./kustomize/invalid_deployment.yaml")

	tt := []struct {
		name      string
		config    *api.KRMInput
		resources []*kyaml.RNode
		expErr    bool
	}{
		{
			name:      "ok config with valid resources",
			config:    okConfig,
			resources: []*kyaml.RNode{okDeploy},
		},
		{
			name:   "ok config, empty resources",
			config: okConfig,
		},
		{
			name:   "invalid config",
			config: invalidConfig,
			expErr: true,
		},
		{
			name:      "ok config with invalid resources",
			config:    okConfig,
			resources: []*kyaml.RNode{invalidDeploy},
			expErr:    true,
		},
		{
			name:      "ok config with both valid and invalid resources",
			config:    okConfig,
			resources: []*kyaml.RNode{okDeploy, invalidDeploy},
			expErr:    true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			provider := model.NewLocalPathProvider("./cue")

			items, err := cuestomize.Cuestomize(
				t.Context(), tc.resources, tc.config, cuestomize.WithModelProvider(provider),
			)
			if tc.expErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// we expect the same number of resources back since this is a validation-only function
				assert.Len(t, items, len(tc.resources))
			}
		})
	}
}

func loadConfig(t *testing.T, config string) *api.KRMInput {
	t.Helper()

	bytes := []byte(config)

	var conf api.KRMInput
	err := yaml.Unmarshal(bytes, &conf)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	return &conf
}

func loadConfigFromFile(t *testing.T, path string) *api.KRMInput {
	t.Helper()

	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config api.KRMInput
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		t.Fatalf("failed to unmarshal config file: %v", err)
	}

	return &config
}

func loadResource(t *testing.T, resourcePath string) *kyaml.RNode {
	t.Helper()

	node, err := kyaml.ReadFile(resourcePath)
	if err != nil {
		t.Fatalf("failed to read resource file: %v", err)
	}

	return node
}
