package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
	"dagger/cuestomize/shared"
	"fmt"
)

const (
	e2eCredSecretContentFmt = `
username=%s
password=%s
`
)

func (m *Cuestomize) UnitTest(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) error {

	// Create a container to run the unit tests
	container := repoBaseContainer(buildContext, nil).
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

func (m *Cuestomize) E2E_Test(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) error {
	// build cuestomize
	cuestomize, err := cuestomizeBuilderContainer(buildContext).Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to build cuestomize: %w", err)
	}
	cuestomizeBinary := cuestomize.File("/workspace/cuestomize")

	testdataDir := buildContext.Directory("e2e/testdata")

	// setup registryNoAuth without authentication
	registryService, err := setupRegistryServiceNoAuth(ctx)
	if err != nil {
		return fmt.Errorf("failed to start registry service: %w", err)
	}
	defer registryService.Stop(ctx)

	// setup registryWithAuth with authentication
	username := "registryuser"
	password := "password"
	registryWithAuthService, err := setupRegistryServiceWithAuth(ctx, username, password)
	if err != nil {
		return fmt.Errorf("failed to start registry with auth service: %w", err)
	}
	defer registryWithAuthService.Stop(ctx)

	// e2e setup (pushing cue module to registries)
	if _, err := testContainerWithRegistryServices(
		buildContext, registryService, registryWithAuthService, username, password).
		WithExec([]string{"go", "run", "./e2e/main.go"}).Sync(ctx); err != nil {
		return fmt.Errorf("failed to run e2e tests: %w", err)
	}

	// run e2e tests
	// TODO: save output to file and extract it for comparison
	kustomize := dag.Container().From(KustomizeImage).
		WithServiceBinding("registry", registryService).
		WithServiceBinding("registry_auth", registryWithAuthService).
		WithDirectory("/testdata", testdataDir).
		WithFile("/bin/cuestomize", cuestomizeBinary).
		WithDirectory("/cue-resources", dag.Directory()).
		WithNewFile(
			"/testdata/kustomize-auth/.env.secret",
			fmt.Sprintf(e2eCredSecretContentFmt, username, password),
		)
	if _, err := kustomize.WithExec([]string{"kustomize", "build", "--enable-alpha-plugins", "--network", "--enable-exec", "/testdata/kustomize"}).Sync(ctx); err != nil {
		return fmt.Errorf("kustomize with no auth e2e failed: %w", err)
	}

	if _, err := kustomize.WithExec([]string{"kustomize", "build", "--enable-alpha-plugins", "--network", "--enable-exec", "/testdata/kustomize-auth"}).Sync(ctx); err != nil {
		return fmt.Errorf("kustomize with auth e2e failed: %w", err)
	}

	return nil
}

func (m *Cuestomize) TestWithCoverage(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) (*dagger.File, error) {
	// Setup registryNoAuth without authentication
	registryService, err := setupRegistryServiceNoAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start registry service: %w", err)
	}
	defer registryService.Stop(ctx)

	// Setup registryWithAuth with authentication
	username := "registryuser"
	password := "password"
	registryWithAuthService, err := setupRegistryServiceWithAuth(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to start registry with auth service: %w", err)
	}
	defer registryWithAuthService.Stop(ctx)

	// Create a container to run the integration tests
	container := testContainerWithRegistryServices(
		buildContext, registryService, registryWithAuthService, username, password).
		WithEnvVariable(shared.IntegrationTestingVarName, "true").
		WithExec([]string{"go", "test", "./...", "-coverprofile=coverage.out"})

	coverageFile := container.File("coverage.out")
	return coverageFile, nil
}

func setupRegistryServiceNoAuth(ctx context.Context) (*dagger.Service, error) {
	registryNoAuth := dag.Container().From(RegistryImage).WithExposedPort(5000)
	return registryNoAuth.AsService().Start(ctx)
}

func setupRegistryServiceWithAuth(ctx context.Context, username, password string) (*dagger.Service, error) {
	htpasswdUtil := dag.Container().From("httpd:2.4").
		WithExec([]string{"htpasswd", "-Bbc", "/tmp/htpasswd", username, password})
	htpasswdFile := htpasswdUtil.File("/tmp/htpasswd")
	registryWithAuth := dag.Container().From(RegistryImage).
		WithFile("/auth/htpasswd", htpasswdFile).
		WithExposedPort(5000).
		WithEnvVariable("REGISTRY_AUTH", "htpasswd").
		WithEnvVariable("REGISTRY_AUTH_HTPASSWD_PATH", "/auth/htpasswd").
		WithEnvVariable("REGISTRY_AUTH_HTPASSWD_REALM", "Dagger Registry")
	return registryWithAuth.AsService().Start(ctx)
}

// testContainerWithRegistryServices returns a repoBaseContainer with registry and registry_auth services bound.
func testContainerWithRegistryServices(buildContext *dagger.Directory, registryService, registryWithAuthService *dagger.Service, username, password string) *dagger.Container {
	return repoBaseContainer(buildContext, nil).
		WithServiceBinding("registry", registryService).
		WithServiceBinding("registry_auth", registryWithAuthService).
		WithEnvVariable(shared.RegistryHostVarName, "registry:5000").
		WithEnvVariable(shared.RegistryAuthHostVarName, "registry_auth:5000").
		WithEnvVariable(shared.RegistryUsernameVarName, username).
		WithEnvVariable(shared.RegistryPasswordVarName, password)
}
