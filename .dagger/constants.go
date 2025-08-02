package main

import "dagger/cuestomize/internal/dagger"

const (
	GolangImage   = "golang:1.24"
	RegistryImage = "registry:2"
)

var (
	DefaultExcludedOpts = dagger.ContainerWithDirectoryOpts{
		Exclude: []string{".dagger/**", ".go-version"},
	}
)
