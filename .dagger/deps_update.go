package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
	"time"
)

func (m *Cuestomize) Renovate(
	ctx context.Context,
	githubToken *dagger.Secret,
	// +default="Workday/cuestomize"
	repo string,
) *dagger.Container {
	cacheHack := time.Now() // avoid dagger to cache the container
	return dag.Container().From("renovate/renovate:latest").
		// WithSecretVariable("RENOVATE_GITHUB_COM_TOKEN", githubToken).
		// WithSecretVariable("GITHUB_COM_TOKEN", githubToken).
		// WithSecretVariable("RENOVATE_GITHUB_TOKEN", githubToken).
		WithSecretVariable("RENOVATE_TOKEN", githubToken). // required
		WithSecretVariable("GITHUB_TOKEN", githubToken).
		WithEnvVariable("RENOVATE_PLATFORM", "github").
		WithEnvVariable("RENOVATE_REQUIRE_CONFIG", "required").
		WithEnvVariable("RENOVATE_DEPENDENCY_DASHBOARD", "false").
		WithEnvVariable("RENOVATE_GIT_AUTHOR", "Renovate Bot <bot@renovateapp.com>").
		WithEnvVariable("RENOVATE_REPOSITORIES", repo).
		WithEnvVariable("CACHE_HACK", cacheHack.String()).
		WithExec([]string{"--dry-run"}, dagger.ContainerWithExecOpts{UseEntrypoint: true})
}
