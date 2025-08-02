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
	builder := repoBaseContainer(ctx, buildContext).
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GO111MODULE", "on").
		WithExec([]string{"go", "build", "-o", "cuestomize", "main.go"})

	// Final stage: create the runtime container with distroless
	container := dag.Container().
		From(DistrolessStaticImage).
		WithFile("/usr/local/bin/cuestomize", builder.File("/workspace/cuestomize")).
		WithEntrypoint([]string{"/usr/local/bin/cuestomize"})

	return container, nil
}

func repoBaseContainer(ctx context.Context, buildContext *dagger.Directory) *dagger.Container {
	// Create a container to run the tests
	return dag.Container().
		From(GolangImage).
		WithWorkdir("/workspace").
		WithFile("/workspace/go.mod", buildContext.File("go.mod")).
		WithFile("/workspace/go.sum", buildContext.File("go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory("/workspace", buildContext, DefaultExcludedOpts)
}
