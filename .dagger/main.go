// A generated module for Cuestomize functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
)

type Cuestomize struct{}

func (m *Cuestomize) Build(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) (*dagger.Container, error) {
	// Build stage: compile the Go binary
	builder := cuestomizeBuilderContainer(buildContext)

	// Final stage: create the runtime container with distroless
	container := dag.Container().
		From(DistrolessStaticImage).
		WithDirectory("/cue-resources", dag.Directory(), dagger.ContainerWithDirectoryOpts{Owner: "nobody"}).
		WithFile("/usr/local/bin/cuestomize", builder.File("/workspace/cuestomize")).
		WithEntrypoint([]string{"/usr/local/bin/cuestomize"})

	return container, nil
}

func (m *Cuestomize) BuildAndPublish(
	ctx context.Context,
	username string,
	password *dagger.Secret,
	// +defaultPath=./
	buildContext *dagger.Directory,
	// +default="ghcr.io"
	registry string,
	repository string,
	tag string,
	// +default=false
	alsoTagAsLatest bool,
) error {
	// lint
	_, err := m.GolangciLintRun(ctx, buildContext, GolangciLintDefaultVersion, "5m")

	// tests
	if err := m.RunTests(ctx, buildContext); err != nil {
		return err
	}

	// Publish stage: push the built image to a registry
	container, err := m.Build(ctx, buildContext)
	if err != nil {
		return err
	}
	container = container.WithRegistryAuth(registry, username, password)

	tags := []string{tag}
	if alsoTagAsLatest {
		tags = append(tags, "latest")
	}
	for _, t := range tags {
		_, err := container.Publish(ctx, registry+"/"+repository+":"+t)
		if err != nil {
			return err
		}
	}

	return nil
}

// repoBaseContainer creates a container with the repository files in it and go dependencies installed.
// The working directory is set to `/workspace` and contains the root of the repository.
func repoBaseContainer(buildContext *dagger.Directory, excludedOpts *dagger.ContainerWithDirectoryOpts) *dagger.Container {
	var exOpts dagger.ContainerWithDirectoryOpts
	if excludedOpts == nil {
		exOpts = DefaultExcludedOpts
	}
	// Create a container to run the tests
	return dag.Container().
		From(GolangImage).
		WithWorkdir("/workspace").
		WithFile("/workspace/go.mod", buildContext.File("go.mod")).
		WithFile("/workspace/go.sum", buildContext.File("go.sum")).
		WithFile("/workspace/.dagger/go.mod", buildContext.File(".dagger/go.mod")).
		WithFile("/workspace/.dagger/go.sum", buildContext.File(".dagger/go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory("/workspace", buildContext, exOpts)
}

// cuestomizeBuilderContainer returns a container that can be used to build the cuestomize binary.
func cuestomizeBuilderContainer(buildContext *dagger.Directory) *dagger.Container {
	return repoBaseContainer(buildContext, nil).
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GO111MODULE", "on").
		WithExec([]string{"go", "build", "-o", "cuestomize", "main.go"})
}
