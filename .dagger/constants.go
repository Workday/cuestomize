package main

import "dagger/cuestomize/internal/dagger"

const (
	GolangImage           = "golang:1.24"
	RegistryImage         = "registry:2"
	DistrolessStaticImage = "gcr.io/distroless/static:latest"
	KustomizeImage        = "registry.k8s.io/kustomize/kustomize:v5.7.1"
)

var (
	DefaultExcludedOpts = dagger.ContainerWithDirectoryOpts{
		Exclude: []string{
			".go-version", "README.md",
			".vscode", "examples/**",
		},
	}
)
