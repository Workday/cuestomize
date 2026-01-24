package main

import (
	"context"
	"dagger/cuestomize/common"
	"dagger/cuestomize/internal/dagger"
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
	// sock *dagger.Socket,
) error {
	// build cuestomize
	cuestomize, err := m.Build(ctx, buildContext, "", "")
	if err != nil {
		return fmt.Errorf("failed to build cuestomize: %w", err)
	}
	cuestomizeBinary := cuestomize.File("/usr/local/bin/cuestomize")

	cuestomizeTar := cuestomize.AsTarball()

	testdataDir := buildContext.Directory("e2e/testdata")

	// setup registryNoAuth without authentication
	registryService, err := setupRegistryServiceNoAuth(ctx)
	if err != nil {
		return fmt.Errorf("failed to start registry service: %w", err)
	}
	defer registryService.Stop(ctx)

	// _, err = dag.Container().WithServiceBinding("registry", registryService).
	// 	Publish(ctx, "registry:5000/cuestomize:latest", dagger.ContainerPublishOpts{
	// 		PlatformVariants: []*dagger.Container{cuestomize},
	// 	})
	// if err != nil {
	// 	return fmt.Errorf("failed to publish cuestomize to registry: %w", err)
	// }

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

	dind := dag.Container().
		From("docker:dind").
		WithEnvVariable("TINI_SUBREAPER", "true").
		WithServiceBinding("registry_auth", registryWithAuthService).
		WithMountedCache("/var/lib/docker", dag.CacheVolume("dind-data")).
		WithExposedPort(2375).AsService(dagger.ContainerAsServiceOpts{
		Args: []string{
			"dockerd", "--tls=false", "--host=tcp://0.0.0.0:2375",
		},
		InsecureRootCapabilities: true,
		UseEntrypoint:            true,
	})

	dindService, err := dind.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start dind: %w", err)
	}
	defer dindService.Stop(ctx)

	dockerCli := dag.Container().From("docker:cli")
	// Load the image into DIND and tag it
	_, err = dockerCli.
		WithServiceBinding("docker-host", dindService).
		WithEnvVariable("DOCKER_HOST", "tcp://docker-host:2375").
		WithFile("/tmp/image.tar", cuestomizeTar).
		WithExec([]string{"sh", "-c", `
		SOURCE=$(docker load -i /tmp/image.tar -q | cut -d' ' -f 4)
		docker tag $SOURCE cuestomize:latest
		`}).Sync(ctx)
	// WithExec([]string{"docker", "load", "-i", "/tmp.image.tar"}).
	// WithExec([]string{"docker", "tag", "", "ghcr.io/workday/cuestomize:latest"}).Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to load cuestomize image into dind: %w", err)
	}

	// run e2e tests
	// TODO: save output to file and extract it for comparison
	kustomize := dag.Container().From(KustomizeImage).
		WithServiceBinding("registry", registryService).
		WithServiceBinding("registry_auth", registryWithAuthService).
		WithServiceBinding("docker-host", dindService).
		WithEnvVariable("DOCKER_HOST", "tcp://docker-host:2375").
		WithDirectory("/testdata", testdataDir).
		WithFile("/bin/cuestomize", cuestomizeBinary).
		WithFile("/usr/local/bin/docker", dockerCli.File("/usr/local/bin/docker")).
		WithDirectory("/cue-resources", dag.Directory()).
		WithNewFile(
			"/testdata/kustomize-auth/.env.secret",
			fmt.Sprintf(e2eCredSecretContentFmt, username, password),
		)
	if _, err := kustomize.WithExec([]string{"kustomize", "build", "--enable-alpha-plugins", "--network", "/testdata/kustomize"}).Sync(ctx); err != nil {
		return fmt.Errorf("kustomize with no auth e2e failed: %w", err)
	}

	if _, err := kustomize.Terminal().WithExec([]string{"kustomize", "build", "--enable-alpha-plugins", "--network", "/testdata/kustomize-auth"}).Sync(ctx); err != nil {
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
		WithEnvVariable(common.IntegrationTestingVarName, "true").
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

func (m *Cuestomize) dind() *dagger.Service {
	return dag.Container().
		From("docker:dind").
		WithEnvVariable("TINI_SUBREAPER", "true").
		WithMountedCache("/var/lib/docker", dag.CacheVolume("dind-data")).
		WithExposedPort(2375).AsService(dagger.ContainerAsServiceOpts{
		Args: []string{
			"dockerd", "--tls=false", "--host=tcp://0.0.0.0:2375",
		},
		InsecureRootCapabilities: true,
		UseEntrypoint:            true,
	})
}

// testContainerWithRegistryServices returns a repoBaseContainer with registry and registry_auth services bound.
func testContainerWithRegistryServices(buildContext *dagger.Directory, registryService, registryWithAuthService *dagger.Service, username, password string) *dagger.Container {
	return repoBaseContainer(buildContext, nil).
		WithServiceBinding("registry", registryService).
		WithServiceBinding("registry_auth", registryWithAuthService).
		WithEnvVariable(common.RegistryHostVarName, "registry:5000").
		WithEnvVariable(common.RegistryAuthHostVarName, "registry_auth:5000").
		WithEnvVariable(common.RegistryUsernameVarName, username).
		WithEnvVariable(common.RegistryPasswordVarName, password)
}
