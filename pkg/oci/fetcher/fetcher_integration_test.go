package fetcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Workday/cuestomize/internal/pkg/testhelpers"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func Test_FetchFromRegistry(t *testing.T) {
	if os.Getenv(testhelpers.IntegrationTestingVarName) != "true" {
		t.Skipf("Skipping test because %s is not set", testhelpers.IntegrationTestingVarName)
	}

	registryNoAuthHost := os.Getenv(testhelpers.RegistryHostVarName)
	if registryNoAuthHost == "" {
		t.Fatalf("Environment variable %s is not set", testhelpers.RegistryHostVarName)
	}
	registryWithAuthHost := os.Getenv(testhelpers.RegistryAuthHostVarName)
	if registryWithAuthHost == "" {
		t.Fatalf("Environment variable %s is not set", testhelpers.RegistryAuthHostVarName)
	}

	username := os.Getenv(testhelpers.RegistryUsernameVarName)
	password := os.Getenv(testhelpers.RegistryPasswordVarName)

	tt := []struct {
		name          string
		testdataDir   string
		registryHost  string
		repo          string
		tag           string
		artifactType  string
		client        remote.Client
		plainHTTP     bool
		shouldError   bool
		expectedFiles []string
	}{
		{
			name:          "fetch from registry without auth",
			testdataDir:   "../../../testdata/integration/sample-module",
			registryHost:  registryNoAuthHost,
			repo:          "sample-module",
			tag:           "latest",
			artifactType:  "application/vnd.cuestomize.module.v1+json",
			plainHTTP:     true,
			expectedFiles: []string{"main.cue", "cue.mod/module.cue"},
		},
		{
			name:         "fetch from registry with auth",
			testdataDir:  "../../../testdata/integration/sample-module",
			registryHost: registryWithAuthHost,
			repo:         "sample-module",
			tag:          "latest",
			artifactType: "application/vnd.cuestomize.module.v1+json",
			client: &auth.Client{
				Credential: auth.StaticCredential(registryWithAuthHost, auth.Credential{
					Username: username,
					Password: password,
				}),
			},
			plainHTTP:     true,
			expectedFiles: []string{"main.cue", "cue.mod/module.cue"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			tempDir := t.TempDir() // Directory to store the fetched artifact

			// push testdata/sample-module to the registry
			_ = testhelpers.PushDirectoryToOCIRegistryT(t, tc.registryHost+"/"+tc.repo+":"+tc.tag, tc.testdataDir, tc.artifactType, tc.tag, tc.client, tc.plainHTTP)

			// Fetch the module from the registry
			err := FetchFromOCIRegistry(ctx, tc.client, tempDir, tc.registryHost, tc.repo, tc.tag, tc.plainHTTP)
			if !tc.shouldError {
				require.NoError(t, err, "failed to fetch module from OCI registry")
				// verify that tempDir contains the expected files
				for _, fileName := range tc.expectedFiles {
					filePath := filepath.Join(tempDir, fileName)
					_, err := os.Stat(filePath)
					require.NoError(t, err, "expected file %s not found in %s", fileName, tempDir)
				}
			} else {
				require.Error(t, err, "expected fetch to error")
			}
		})
	}
}
