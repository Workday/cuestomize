package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
	"fmt"

	"dagger.io/dagger/dag"
)

func (m *Cuestomize) GolangciLintRun(
	ctx context.Context,
	// +defaultPath=./
	src *dagger.Directory,
	// +optional
	version string,
	// +default="5m"
	timeout string,
) (*dagger.Container, error) {
	image := GolangciLintImage
	if version != "" {
		image = fmt.Sprintf(GolangciLintImageFmt, version)
	}
	linter := dag.Container().From(image).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src")

	return linter.WithExec([]string{
		"golangci-lint",
		"run",
		"-v",
		fmt.Sprintf("--timeout=%s", timeout),
	}).Sync(ctx)
}
