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
	container := getRepoTestContainer(ctx, buildContext).
		WithExec([]string{"go", "test", "./..."})

	out, err := container.Stdout(ctx)
	fmt.Printf(out)
	return err
}

func (m *Cuestomize) IntegrationTest(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) error {

	// Setup registryNoAuth without authentication
	registryNoAuth := dag.Container().From(RegistryImage).
		WithExposedPort(5000)
	registryService, err := registryNoAuth.AsService().WithHostname("registry.noauth.local").Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start registry service: %w", err)
	}
	defer registryService.Stop(ctx)

	// Create a container to run the integration tests
	out, err := dag.Container().
		From(GolangImage).
		WithServiceBinding("registry", registryService).
		WithWorkdir("/workspace").
		WithFile("/workspace/go.mod", buildContext.File("go.mod")).
		WithFile("/workspace/go.sum", buildContext.File("go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory("/workspace", buildContext, DefaultExcludedOpts).
		WithEnvVariable("INTEGRATION_TEST", "true").
		WithEnvVariable("REGISTRY_HOST", "registry:5000").
		WithExec([]string{"go", "test", "./integration"}).Stdout(ctx)

	fmt.Printf(out)

	return err
}

func getRepoTestContainer(ctx context.Context, buildContext *dagger.Directory) *dagger.Container {
	// Create a container to run the tests
	return dag.Container().
		From(GolangImage).
		WithWorkdir("/workspace").
		WithFile("/workspace/go.mod", buildContext.File("go.mod")).
		WithFile("/workspace/go.sum", buildContext.File("go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory("/workspace", buildContext, DefaultExcludedOpts)
}
