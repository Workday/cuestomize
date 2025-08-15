package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
)

const (
	CueModuleArtifactType     = "application/vnd.cue.module.v1+json"
	CueModuleFileArtifactType = "application/vnd.cue.modulefile.v1"
)

func (m *Cuestomize) PublishExamples(
	ctx context.Context,
	username string,
	password *dagger.Secret,
	// +defaultPath=./
	buildContext *dagger.Directory,
	// +default="ghcr.io"
	registry string,
	// +default="workday/cuestomize/cuemodules/cuestomize-examples"
	repositoryPrefix string,
	// +default="latest"
	tag string,
	// +optional
	latest bool,
	// +default="info"
	logLevel string,
) *dagger.File {
	container := m.GoGenerate(ctx, buildContext)

	latestStr := "false"
	if latest {
		latestStr = "true"
	}
	container = container.
		WithEnvVariable("LOG_LEVEL", logLevel).
		WithEnvVariable("OCI_REGISTRY", registry).
		WithEnvVariable("OCI_REPOSITORY_PREFIX", repositoryPrefix).
		WithEnvVariable("IS_LATEST", latestStr).
		WithEnvVariable("OCI_USERNAME", username).
		WithSecretVariable("OCI_PASSWORD", password).
		WithExec([]string{"go", "run", "hack/push-examples.go", tag}, dagger.ContainerWithExecOpts{
			RedirectStderr: "stderr.log",
		})

	file := container.File("stderr.log")
	return file
}
