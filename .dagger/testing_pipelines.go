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

	// Create a container to run the integration tests
	exitCode, err := repoBaseContainer(buildContext).
		WithEnvVariable("INTEGRATION_TEST", "true").
		WithEnvVariable("REGISTRY_HOST", "registry:5000").
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
