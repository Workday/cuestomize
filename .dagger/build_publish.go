package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
	"fmt"
)

func (m *Cuestomize) Build(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
	// +default=""
	platform string,
	// +default=""
	ldflags string,
) *dagger.Container {
	containerOpts := dagger.ContainerOpts{}
	if platform != "" {
		containerOpts.Platform = dagger.Platform(platform)
	}

	container := buildContext.DockerBuild(dagger.DirectoryDockerBuildOpts{
		Dockerfile: "Containerfile",
	})

	return container
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
	// +default="-s -w"
	ldflags string,
	tag string,
	// +default=false
	alsoTagAsLatest bool,
	// +default=true
	runValidations bool,
	// +default=[]
	platforms []string,
) error {
	if runValidations {
		// lint
		if _, err := m.GolangciLintRun(ctx, buildContext, "", "5m"); err != nil {
			return err
		}

		// tests
		if _, err := m.TestWithCoverage(ctx, buildContext); err != nil {
			return err
		}
	}

	if len(platforms) == 0 {
		platform, err := dag.DefaultPlatform(ctx)
		if err != nil {
			return err
		}
		platforms = append(platforms, string(platform))
	}

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	for _, platform := range platforms {
		container := m.Build(ctx, buildContext, string(platform), ldflags)
		platformVariants = append(platformVariants, container)
	}

	tags := []string{tag}
	if alsoTagAsLatest {
		tags = append(tags, "latest")
	}
	for _, t := range tags {
		_, err := dag.Container().WithRegistryAuth(registry, username, password).
			Publish(ctx, fmt.Sprintf("%v/%v:%v", registry, repository, t), dagger.ContainerPublishOpts{
				PlatformVariants: platformVariants,
			})
		if err != nil {
			return err
		}
	}

	return nil
}
