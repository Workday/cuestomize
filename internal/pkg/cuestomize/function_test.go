package cuestomize

import (
	"testing"

	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/internal/pkg/testhelpers"
	"sigs.k8s.io/kustomize/kyaml/resid"

	"github.com/stretchr/testify/require"
)

const (
	// KrmFunPath is the path, inside a given testdata directory, where the KRM function YAML is stored.
	KrmFunPath = "krm-func.yaml"
	// ItemsPath is the path, inside a given testdata directory, where the items YAML is stored.
	ItemsPath = "items.yaml"
)

func TestKRMFunction(t *testing.T) {
	tests := []struct {
		Name                  string
		TestdataCUEModelPath  string
		TestdataKustomizePath string
		ShouldFail            bool
		Expected              []resid.ResId
	}{
		// configmap-model tests
		{
			Name:                  "configmap-model with configmap-ok should succeed",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/configmap-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/configmap-ok",
			ShouldFail:            false,
			Expected: []resid.ResId{
				resid.NewResIdWithNamespace(resid.Gvk{Group: "cuestomize.dev", Version: "v1alpha1", Kind: "Cuestomization"}, "example-cuestomization", ""),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}, "example-deployment", "example-namespace"),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "", Version: "v1", Kind: "Service"}, "example-service", "example-namespace"),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "", Version: "v1", Kind: "ConfigMap"}, "example-configmap", "default"),
			},
		},
		// configmap-struct-model tests
		{
			Name:                  "configmap-struct-model with configmap-ok should succeed",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/configmap-struct-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/configmap-ok",
			ShouldFail:            false,
			Expected: []resid.ResId{
				resid.NewResIdWithNamespace(resid.Gvk{Group: "cuestomize.dev", Version: "v1alpha1", Kind: "Cuestomization"}, "example-cuestomization", ""),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}, "example-deployment", "example-namespace"),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "", Version: "v1", Kind: "Service"}, "example-service", "example-namespace"),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "", Version: "v1", Kind: "ConfigMap"}, "example-configmap", "default"),
			},
		},
		{
			Name:                  "configmap-model with configmap-missing-includes should fail",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/configmap-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/configmap-missing-includes",
			ShouldFail:            true,
		},
		{
			Name:                  "configmap-model with configmap-wrong-apiversion should fail",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/configmap-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/configmap-wrong-apiversion",
			ShouldFail:            true,
		},
		// deployment-model tests
		{
			Name:                  "deployment-model with deployment-ok should succeed",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/deployment-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/deployment-ok",
			ShouldFail:            false,
			Expected: []resid.ResId{
				resid.NewResIdWithNamespace(resid.Gvk{Group: "cuestomize.dev", Version: "v1alpha1", Kind: "NginxDeployment"}, "nginx-deployment", ""),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "", Version: "v1", Kind: "Service"}, "example-service", "example-namespace"),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}, "example-deployment", "example-namespace"),
				resid.NewResIdWithNamespace(resid.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}, "nginx-deployment", "nginx"),
			},
		},
		{
			Name:                  "deployment-model with deployment-unexpected-includes should fail",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/deployment-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/deployment-unexpected-includes",
			ShouldFail:            true,
		},
		{
			Name:                  "deployment-model with deployment-wrong-input-shape should fail",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/deployment-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/deployment-wrong-input-shape",
			ShouldFail:            true,
		},
		{
			Name:                  "deployment-model with deployment-wrong-kind should fail",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/deployment-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/deployment-wrong-kind",
			ShouldFail:            true,
		},
		// fuzzy-model tests
		{
			Name:                  "configmap-model with deployment-ok should fail",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/configmap-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/deployment-ok",
			ShouldFail:            true,
		},
		{
			Name:                  "deployment-model with configmap-ok should fail",
			TestdataCUEModelPath:  "../../../testdata/function/cue-modules/deployment-model",
			TestdataKustomizePath: "../../../testdata/function/kustomize-inputs/configmap-ok",
			ShouldFail:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			krmFuncPath := tt.TestdataKustomizePath + "/" + KrmFunPath
			itemsPath := tt.TestdataKustomizePath + "/" + ItemsPath

			config := testhelpers.LoadFromFile[api.KRMInput](t, krmFuncPath)

			items := testhelpers.LoadResourceList(t, krmFuncPath, itemsPath)

			krmFunc, err := NewBuilder().SetResourcesPath(tt.TestdataCUEModelPath).SetConfig(config).Build()
			require.NoError(t, err, "KRMFuncBuilder failed to build KRM function")

			result, err := krmFunc(items)
			if tt.ShouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				for _, res := range result {
					resId := resid.FromRNode(res)
					require.Contains(t, tt.Expected, resId, "Expected resource ID not found in result")
				}
			}

		})
	}
}
