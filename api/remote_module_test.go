package api

import (
	"testing"
)

func TestRemoteModule_GetRegistry(t *testing.T) {
	tests := []struct {
		name    string
		module  RemoteModule
		want    string
		wantErr bool
	}{
		{
			name: "uses ref when present",
			module: RemoteModule{
				Ref:      "ghcr.io/workday/module:v1.0.0",
				Registry: "old-registry.io",
			},
			want:    "ghcr.io",
			wantErr: false,
		},
		{
			name: "falls back to registry field when ref is empty",
			module: RemoteModule{
				Registry: "docker.io",
			},
			want:    "docker.io",
			wantErr: false,
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
			got, err := tt.module.GetRegistry()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteModule_GetRepo(t *testing.T) {
	tests := []struct {
		name    string
		module  RemoteModule
		want    string
		wantErr bool
	}{
		{
			name: "uses ref when present",
			module: RemoteModule{
				Ref:  "ghcr.io/workday/my-module:v1.0.0",
				Repo: "old-repo",
			},
			want:    "workday/my-module",
			wantErr: false,
		},
		{
			name: "falls back to repo field when ref is empty",
			module: RemoteModule{
				Repo: "library/nginx",
			},
			want:    "library/nginx",
			wantErr: false,
		},
		{
			name: "handles nested repo paths",
			module: RemoteModule{
				Ref: "registry.io/org/team/project:v2.0.0",
			},
			want:    "org/team/project",
			wantErr: false,
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
			got, err := tt.module.GetRepo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteModule_GetTag(t *testing.T) {
	tests := []struct {
		name    string
		module  RemoteModule
		want    string
		wantErr bool
	}{
		{
			name: "uses ref when present",
			module: RemoteModule{
				Ref: "ghcr.io/workday/module:v1.0.0",
				Tag: "old-tag",
			},
			want:    "v1.0.0",
			wantErr: false,
		},
		{
			name: "falls back to tag field when ref is empty",
			module: RemoteModule{
				Tag: "v2.0.0",
			},
			want:    "v2.0.0",
			wantErr: false,
		},
		{
			name: "defaults to latest when no tag in ref",
			module: RemoteModule{
				Ref: "ghcr.io/workday/module",
			},
			want:    "latest",
			wantErr: false,
		},
		{
			name: "handles sha256 tags",
			module: RemoteModule{
				Ref: "docker.io/nginx:sha256-abc123",
			},
			want:    "sha256-abc123",
			wantErr: false,
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
			got, err := tt.module.GetTag()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetTag() = %v, want %v", got, tt.want)
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

		registry, err := module.GetRegistry()
		if err != nil {
			t.Errorf("GetRegistry() unexpected error = %v", err)
		}
		if registry != "ghcr.io" {
			t.Errorf("GetRegistry() = %v, want ghcr.io", registry)
		}

		repo, err := module.GetRepo()
		if err != nil {
			t.Errorf("GetRepo() unexpected error = %v", err)
		}
		if repo != "workday/module" {
			t.Errorf("GetRepo() = %v, want workday/module", repo)
		}

		tag, err := module.GetTag()
		if err != nil {
			t.Errorf("GetTag() unexpected error = %v", err)
		}
		if tag != "v1.0.0" {
			t.Errorf("GetTag() = %v, want v1.0.0", tag)
		}
	})

	t.Run("ref takes precedence over deprecated fields", func(t *testing.T) {
		module := RemoteModule{
			Ref:      "new-registry.io/new-repo:v2.0.0",
			Registry: "old-registry.io",
			Repo:     "old-repo",
			Tag:      "v1.0.0",
		}

		registry, _ := module.GetRegistry()
		if registry != "new-registry.io" {
			t.Errorf("GetRegistry() = %v, want new-registry.io (ref should take precedence)", registry)
		}

		repo, _ := module.GetRepo()
		if repo != "new-repo" {
			t.Errorf("GetRepo() = %v, want new-repo (ref should take precedence)", repo)
		}

		tag, _ := module.GetTag()
		if tag != "v2.0.0" {
			t.Errorf("GetTag() = %v, want v2.0.0 (ref should take precedence)", tag)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
