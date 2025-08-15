package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
)

// GoGenerate runs go generate inside the provided build context.
func (m *Cuestomize) GoGenerate(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
) *dagger.Container {
	cue := dag.Container().From(CuelangImage)

	container := repoBaseContainer(buildContext, nil).
		WithFile(
			"/usr/bin/cue",
			cue.File("/usr/bin/cue"),
		).
		WithExec([]string{"go", "generate", "./..."})
	return container
}
