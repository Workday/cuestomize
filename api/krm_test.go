package api

import (
	"testing"

	"github.com/Workday/cuestomize/internal/pkg/testloaders"
	"github.com/stretchr/testify/require"
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
			krmInput := testloaders.LoadFromFile[KRMInput](t, tt.testdataDir+"/"+TestKRMInputFileName)
			items := testloaders.LoadResourceList(t, tt.testdataDir+"/"+TestKRMInputFileName, tt.testdataDir+"/"+TestItemsFileName)

			includes, err := ExtractIncludes(krmInput, items)

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
