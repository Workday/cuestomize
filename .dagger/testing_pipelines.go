package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
	"fmt"
)

func (m *Cuestomize) UnitTest(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) error {

	// Create a container to run the unit tests
	container := repoBaseContainer(buildContext).
		WithExec([]string{"go", "test", "./..."})

	exitCode, err := container.ExitCode(ctx)
	if err != nil {
		return fmt.Errorf("failed to run unit tests: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("unit tests failed with exit code %d", exitCode)
	}
	return nil
}

func (m *Cuestomize) IntegrationTest(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) error {

	// Setup registryNoAuth without authentication
	registryNoAuth := dag.Container().From(RegistryImage).WithExposedPort(5000)
	registryService, err := registryNoAuth.AsService().Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start registry service: %w", err)
	}
	defer registryService.Stop(ctx)

	// Setup registryWithAuth with authentication
	username := "registryuser"
	password := "password"

	htpasswdUtil := dag.Container().From("httpd:2.4").
		WithExec([]string{"htpasswd", "-Bbc", "/tmp/htpasswd", username, password})
	htpasswdFile := htpasswdUtil.File("/tmp/htpasswd")
	registryWithAuth := dag.Container().From(RegistryImage).
		WithFile("/auth/htpasswd", htpasswdFile).
		WithExposedPort(5000).
		WithEnvVariable("REGISTRY_AUTH", "htpasswd").
		WithEnvVariable("REGISTRY_AUTH_HTPASSWD_PATH", "/auth/htpasswd").
		WithEnvVariable("REGISTRY_AUTH_HTPASSWD_REALM", "Dagger Registry")
	registryWithAuthService, err := registryWithAuth.AsService().Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start registry with auth service: %w", err)
	}
	defer registryWithAuthService.Stop(ctx)

	// Create a container to run the integration tests
	exitCode, err := repoBaseContainer(buildContext).
		WithServiceBinding("registry", registryService).
		WithServiceBinding("registry_auth", registryWithAuthService).
		WithEnvVariable("INTEGRATION_TEST", "true").
		WithEnvVariable("REGISTRY_HOST", "registry:5000").
		WithEnvVariable("REGISTRY_AUTH_HOST", "registry_auth:5000").
		WithEnvVariable("REGISTRY_USERNAME", username).
		WithEnvVariable("REGISTRY_PASSWORD", password).
		WithExec([]string{"go", "test", "./integration"}).ExitCode(ctx)

	if err != nil {
		return fmt.Errorf("failed to run integration tests: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("integration tests failed with exit code %d", exitCode)
	}
	return nil
}

func (m *Cuestomize) RunTests(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) error {
	if err := m.UnitTest(ctx, buildContext); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}
	if err := m.IntegrationTest(ctx, buildContext); err != nil {
		return fmt.Errorf("integration tests failed: %w", err)
	}
	return nil
}
