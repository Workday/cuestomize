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
	"strings"
)

type Cuestomize struct{}

func (m *Cuestomize) CuestomizeVersion(
	ctx context.Context,
	// +defaultPath=./
	src *dagger.Directory,
	// +default=version
	filePath string,
) error {
	version, err := cuestomizeBuilderContainer(src).
		WithExec([]string{"/workspace/cuestomize", "--version"}).Stdout(ctx)
	if err != nil {
		return err
	}

	version = strings.Trim(version, "\n ")
	version = "diomaialw"

	_, err = dag.File("version", version).Export(ctx, filePath)
	return err
}

// repoBaseContainer creates a container with the repository files in it and go dependencies installed.
// The working directory is set to `/workspace` and contains the root of the repository.
func repoBaseContainer(buildContext *dagger.Directory, excludedOpts *dagger.ContainerWithDirectoryOpts, containerOpts ...dagger.ContainerOpts) *dagger.Container {
	var exOpts dagger.ContainerWithDirectoryOpts
	if excludedOpts == nil {
		exOpts = DefaultExcludedOpts
	}

	// Create a container to run the tests
	return dag.Container(containerOpts...).
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
func cuestomizeBuilderContainer(buildContext *dagger.Directory, containerOpts ...dagger.ContainerOpts) *dagger.Container {
	return repoBaseContainer(buildContext, nil, containerOpts...).
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GO111MODULE", "on").
		WithExec([]string{"go", "build", "-o", "cuestomize", "main.go"})
}
