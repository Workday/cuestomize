package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	semVerRegexp = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)
)

// Builds the Cuestomize image.
func (m *Cuestomize) Build(
	ctx context.Context,
	// +defaultPath=./
	buildContext *dagger.Directory,
	// +defaultPath="./.git"
	git *dagger.GitRepository,
	// +default=""
	platform string,
	// +default=""
	ldflags string,
	// +default="nightly"
	version string,
) *dagger.Container {
	containerOpts := dagger.ContainerOpts{}
	if platform != "" {
		containerOpts.Platform = dagger.Platform(platform)
	}
	builder := cuestomizeBuilderContainer(buildContext, ldflags, containerOpts)

	ldflags = fmt.Sprintf("-X 'main.Version=%s' %s", version, ldflags)

	commit, err := git.Head().Commit(ctx)
	if err != nil {
		panic("failed to get git commit: " + err.Error())
	}

	container := dag.Container(containerOpts).
		From(DistrolessStaticImage).
		WithDirectory("/cue-resources", dag.Directory(), dagger.ContainerWithDirectoryOpts{Owner: "nobody"}).
		WithFile("/usr/local/bin/cuestomize", builder.File("/workspace/cuestomize")).
		WithEntrypoint([]string{"/usr/local/bin/cuestomize"}).
		WithAnnotation("org.opencontainers.image.title", "Cuestomize").
		WithAnnotation("org.opencontainers.image.description", "KRM function integrating CUE with Kustomize for Kubernetes configuration management").
		WithAnnotation("org.opencontainers.image.documentation", "https://workday.github.io/cuestomize").
		WithAnnotation("org.opencontainers.image.created", time.Now().UTC().Format(time.RFC3339)).
		WithAnnotation("org.opencontainers.image.source", "https://github.com/Workday/cuestomize").
		WithAnnotation("org.opencontainers.image.version", strings.TrimPrefix(version, "v")).
		WithAnnotation("org.opencontainers.image.revision", commit)

	return container
}

// Builds the Cuestomize binary for the specified platforms and publishes it to the specified container registry with appropriate tags.
// The version string is parsed to generate additional tags for standard semantic versioning formats. For example, a version "1.2.3" would generate tags "1.2.3", "1", and "1.2".
// If the version string includes pre-release or build metadata (e.g., "1.2.3-alpha"), only the full version string will be used as a tag.
func (m *Cuestomize) BuildAndPublish(
	ctx context.Context,
	username string,
	password *dagger.Secret,
	// +defaultPath=./
	buildContext *dagger.Directory,
	// +defaultPath="./.git"
	git *dagger.GitRepository,
	// +default="ghcr.io"
	registry string,
	repository string,
	// +default="-s -w"
	ldflags string,
	// +default="nightly"
	version string,
	// +default=false
	latest bool,
	// +default=[]
	platforms []string,
) error {
	if len(platforms) == 0 {
		platform, err := dag.DefaultPlatform(ctx)
		if err != nil {
			return err
		}
		platforms = append(platforms, string(platform))
	}

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	for _, platform := range platforms {
		container := m.Build(ctx, buildContext, git, string(platform), ldflags, version)
		platformVariants = append(platformVariants, container)
	}

	tags := versionToTags(version)
	if latest {
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

func versionToTags(version string) []string {
	tags := []string{version}

	matches := semVerRegexp.FindStringSubmatch(version)

	if matches == nil {
		// version doesn't match semver pattern
		return tags
	}

	if matches[len(matches)-1] != "" {
		// version has pre-release or build metadata, don't generate additional tags
		return tags
	}

	tag := matches[1]
	tags = append(tags, tag)

	tag = fmt.Sprintf("%s.%s", tag, matches[2])
	tags = append(tags, tag)

	return tags
}
