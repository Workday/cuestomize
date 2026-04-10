package main

import "dagger/cuestomize/internal/dagger"

// Note: when updating these constants, also update renovate.json5
// as they are updated in there through regexes.
const (
	// GolangImage is the Golang base image
	GolangImage = "golang:1.26"
	// RegistryImage is image for local container registry
	RegistryImage = "registry:3"
	// DistrolessStaticImage is the distroless static image
	DistrolessStaticImage = "gcr.io/distroless/static:latest"
	// KustomizeImage is the Kustomize image
	KustomizeImage = "registry.k8s.io/kustomize/kustomize:v5.8.1"
	// CuelangVersion is the version of Cuelang
	CuelangVersion = "v0.16.0"
	// GolangciLintImage is the GolangCI-Lint image used by default
	GolangciLintImage = "golangci/golangci-lint:v2.11.4-alpine"
	// GitImage is the image used for Git operations in Dagger
	GitImage = "alpine/git:2.52.0"
)

const (
	// GolangciLintImageFmt is the format for the GolangCI-Lint image. It accepts the version as a string
	GolangciLintImageFmt = "golangci/golangci-lint:%s-alpine"
)

var (
	DefaultExcludedOpts = dagger.ContainerWithDirectoryOpts{
		Exclude: []string{
			".go-version", "README.md",
			".vscode", ".dagger", "docs",
		},
		Gitignore: true,
	}
)
