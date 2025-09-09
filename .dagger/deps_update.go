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
	// +default=false
	dryRun bool,
) *dagger.Container {
	args := []string{}
	if dryRun {
		args = append(args, "--dry-run")
	}

	cacheHack := time.Now() // avoid dagger to cache the container
	return dag.Container().From("renovate/renovate:latest").
		WithSecretVariable("RENOVATE_TOKEN", githubToken).
		WithEnvVariable("RENOVATE_PLATFORM", "github").
		WithEnvVariable("RENOVATE_REQUIRE_CONFIG", "required").
		WithEnvVariable("RENOVATE_DEPENDENCY_DASHBOARD", "false").
		WithEnvVariable("RENOVATE_GIT_AUTHOR", "Renovate Bot <bot@renovateapp.com>").
		WithEnvVariable("RENOVATE_REPOSITORIES", repo).
		WithEnvVariable("CACHE_HACK", cacheHack.String()).
		WithExec(args, dagger.ContainerWithExecOpts{UseEntrypoint: true})
}
