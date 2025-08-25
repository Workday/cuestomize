package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
	"fmt"
)

// GoGenerate runs go generate inside the provided build context.
func (m *Cuestomize) GoGenerate(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) *dagger.Container {
	container := repoBaseContainer(buildContext, &dagger.ContainerWithDirectoryOpts{
		Exclude: []string{
			".go-version", "README.md", ".vscode",
		},
	}).
		WithExec([]string{"go", "install", fmt.Sprintf("cuelang.org/go/cmd/cue@%s", CuelangVersion)}).
		WithExec([]string{"go", "generate", "./..."})
	return container
}
