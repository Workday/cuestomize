package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
)

func (m *Cuestomize) Build(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
	// +default=""
	platform string,
	// +default=""
	ldflags string,
) (*dagger.Container, error) {
	containerOpts := dagger.ContainerOpts{}
	if platform != "" {
		containerOpts.Platform = dagger.Platform(platform)
	}

	// Build stage: compile the Go binary
	builder := cuestomizeBuilderContainer(buildContext, ldflags, containerOpts)

	// Final stage: create the runtime container with distroless
	container := dag.Container(containerOpts).
		From(DistrolessStaticImage).
		WithDirectory("/cue-resources", dag.Directory(), dagger.ContainerWithDirectoryOpts{Owner: "nobody"}).
		WithFile("/usr/local/bin/cuestomize", builder.File("/workspace/cuestomize")).
		WithEntrypoint([]string{"/usr/local/bin/cuestomize"})

	return container, nil
}

func (m *Cuestomize) Save(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
	// +default=""
	platform string,
	// +default=""
	ldflags string) *dagger.File {
	container, _ := m.Build(ctx, buildContext, string(platform), ldflags)

	return container.AsTarball()
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
		container, err := m.Build(ctx, buildContext, string(platform), ldflags)
		if err != nil {
			return err
		}
		platformVariants = append(platformVariants, container)
	}

	tags := []string{tag}
	if alsoTagAsLatest {
		tags = append(tags, "latest")
	}
	for _, t := range tags {
		_, err := dag.Container().WithRegistryAuth(registry, username, password).
			Publish(ctx, registry+"/"+repository+":"+t, dagger.ContainerPublishOpts{
				PlatformVariants: platformVariants,
			})
		if err != nil {
			return err
		}
	}

	return nil
}
