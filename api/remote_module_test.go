package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteModule_GetReference(t *testing.T) {
	tests := []struct {
		name         string
		module       RemoteModule
		wantRegistry string
		wantRepo     string
		wantRef      string
		wantErr      bool
	}{
		{
			name: "uses ref when present",
			module: RemoteModule{
				Ref:      "ghcr.io/workday/module:v1.0.0",
				Registry: "old-registry.io",
				Repo:     "old-repo",
				Tag:      "old-tag",
			},
			wantRegistry: "ghcr.io",
			wantRepo:     "workday/module",
			wantRef:      "v1.0.0",
			wantErr:      false,
		},
		{
			name: "constructs from separate fields when ref is empty",
			module: RemoteModule{
				Registry: "docker.io",
				Repo:     "library/nginx",
				Tag:      "latest",
			},
			wantRegistry: "docker.io",
			wantRepo:     "library/nginx",
			wantRef:      "latest",
			wantErr:      false,
		},
		{
			name: "handles nested repo paths",
			module: RemoteModule{
				Ref: "registry.io/org/team/project:v2.0.0",
			},
			wantRegistry: "registry.io",
			wantRepo:     "org/team/project",
			wantRef:      "v2.0.0",
			wantErr:      false,
		},
		{
			name: "handles registry with port",
			module: RemoteModule{
				Ref: "registry:5000/sample-module:latest",
			},
			wantRegistry: "registry:5000",
			wantRepo:     "sample-module",
			wantRef:      "latest",
			wantErr:      false,
		},
		{
			name: "handles localhost registry",
			module: RemoteModule{
				Ref: "localhost:5000/my-module:v1.0.0",
			},
			wantRegistry: "localhost:5000",
			wantRepo:     "my-module",
			wantRef:      "v1.0.0",
			wantErr:      false,
		},
		{
			name: "constructs without tag defaults to empty",
			module: RemoteModule{
				Registry: "ghcr.io",
				Repo:     "workday/module",
			},
			wantRegistry: "ghcr.io",
			wantRepo:     "workday/module",
			wantRef:      "",
			wantErr:      false,
		},
		{
			name: "handles sha256 tags",
			module: RemoteModule{
				Ref: "docker.io/nginx:sha256-abc123",
			},
			wantRegistry: "docker.io",
			wantRepo:     "nginx",
			wantRef:      "sha256-abc123",
			wantErr:      false,
		},
		{
			name: "returns error for invalid ref",
			module: RemoteModule{
				Ref: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := tt.module.GetReference()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.Equal(t, tt.wantRegistry, ref.Registry)
				assert.Equal(t, tt.wantRepo, ref.Repository)
				assert.Equal(t, tt.wantRef, ref.Reference)
			}
		})
	}
}

func TestRemoteModule_BackwardsCompatibility(t *testing.T) {
	t.Run("deprecated fields still work", func(t *testing.T) {
		module := RemoteModule{
			Registry: "ghcr.io",
			Repo:     "workday/module",
			Tag:      "v1.0.0",
		}

		ref, err := module.GetReference()
		require.NoError(t, err, "GetReference() unexpected error = %v", err)

		assert.Equal(t, "ghcr.io", ref.Registry)
		assert.Equal(t, "workday/module", ref.Repository)
		assert.Equal(t, "v1.0.0", ref.Reference)
	})

	t.Run("ref takes precedence over deprecated fields", func(t *testing.T) {
		module := RemoteModule{
			Ref:      "new-registry.io/new-repo:v2.0.0",
			Registry: "old-registry.io",
			Repo:     "old-repo",
			Tag:      "v1.0.0",
		}

		ref, err := module.GetReference()
		require.NoError(t, err)

		assert.Equal(t, "new-registry.io", ref.Registry)
		assert.Equal(t, "new-repo", ref.Repository)
		assert.Equal(t, "v2.0.0", ref.Reference)
	})
}
