// Cuestomize CI/CD functions

package main

import (
	"dagger/cuestomize/internal/dagger"
)

type Cuestomize struct{}

// repoBaseContainer creates a container with the repository files in it and go dependencies installed.
// The working directory is set to `/workspace` and contains the root of the repository.
func repoBaseContainer(buildContext *dagger.Directory, dirOpts *dagger.ContainerWithDirectoryOpts, containerOpts ...dagger.ContainerOpts) *dagger.Container {
	if dirOpts == nil {
		dirOpts = &DefaultExcludedOpts
	}

	// Create a container to run the tests
	return dag.Container(containerOpts...).
		From(GolangImage).
		WithWorkdir("/workspace").
		WithFile("/workspace/go.mod", buildContext.File("go.mod")).
		WithFile("/workspace/go.sum", buildContext.File("go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory("/workspace", buildContext, *dirOpts)
}
