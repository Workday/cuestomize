package main

import "dagger/cuestomize/internal/dagger"

const (
	GolangImage           = "golang:1.25"
	RegistryImage         = "registry:2"
	DistrolessStaticImage = "gcr.io/distroless/static:latest"
	KustomizeImage        = "registry.k8s.io/kustomize/kustomize:v5.7.1"
	CuelangVersion        = "v0.14.1"

	GolangciLintDefaultVersion = "v2.4.0"
	GolangciLingImageFmt       = "golangci/golangci-lint:%s-alpine"
)

var (
	DefaultExcludedOpts = dagger.ContainerWithDirectoryOpts{
		Exclude: []string{
			".go-version", "README.md",
			".vscode", "examples",
		},
	}
)
