package main

import "dagger/cuestomize/internal/dagger"

const (
	GolangImage           = "golang:1.24"
	RegistryImage         = "registry:2"
	DistrolessStaticImage = "gcr.io/distroless/static:latest"
)

var (
	DefaultExcludedOpts = dagger.ContainerWithDirectoryOpts{
		Exclude: []string{".dagger/**", ".go-version"},
	}
)
