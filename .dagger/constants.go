package main

import "dagger/cuestomize/internal/dagger"

// Note: when updating these constants, also update renovate.json5
// as they are updated in there through regexes.
const (
	// GolangImage is the Golang base image
	GolangImage = "golang:1.25"
	// RegistryImage is image for local container registry
	RegistryImage = "registry:3"
	// DistrolessStaticImage is the distroless static image
	DistrolessStaticImage = "gcr.io/distroless/static:latest"
	// KustomizeImage is the Kustomize image
	KustomizeImage = "registry.k8s.io/kustomize/kustomize:v5.7.1"
	// CuelangVersion is the version of Cuelang
	CuelangVersion = "v0.15.1"
	// GolangciLintImage is the GolangCI-Lint image used by default
	GolangciLintImage = "golangci/golangci-lint:v2.7.2-alpine"
)

const (
	// GolangciLintImageFmt is the format for the GolangCI-Lint image. It accepts the version as a string
	GolangciLintImageFmt = "golangci/golangci-lint:%s-alpine"
)

var (
	DefaultExcludedOpts = dagger.ContainerWithDirectoryOpts{
		Exclude: []string{
			".go-version", "README.md",
			".vscode", "examples",
		},
	}
)
