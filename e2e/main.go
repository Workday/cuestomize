// Package main provides an e2e test setup utility that uploads CUE modules to local registries.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Workday/cuestomize/internal/pkg/testhelpers"
	"github.com/Workday/cuestomize/pkg/oci"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// main uploads e2e/testdata/cue module to local registry with and without authentication
// for e2e to run.
func main() {
	ctx := context.Background()

	registryNoAuthHost := os.Getenv(testhelpers.RegistryHostVarName)
	if registryNoAuthHost == "" {
		panic(fmt.Sprintf("Environment variable %s is not set", testhelpers.RegistryHostVarName))
	}
	registryWithAuthHost := os.Getenv(testhelpers.RegistryAuthHostVarName)
	if registryWithAuthHost == "" {
		panic(fmt.Sprintf("Environment variable %s is not set", testhelpers.RegistryAuthHostVarName))
	}

	username := os.Getenv(testhelpers.RegistryUsernameVarName)
	password := os.Getenv(testhelpers.RegistryPasswordVarName)

	tag := "latest"
	artifactType := "application/vnd.cuestomize.module.v1+json"

	plainHTTP := true

	// push to registry with no authentication
	if _, err := oci.PushDirectoryToOCIRegistry(ctx, registryNoAuthHost+"/sample-module:"+tag, "e2e/testdata/cue", artifactType, tag, nil, plainHTTP); err != nil {
		panic(err)
	}

	// push to registry with authentication
	if _, err := oci.PushDirectoryToOCIRegistry(ctx, registryWithAuthHost+"/sample-module:"+tag, "e2e/testdata/cue", artifactType, tag, &auth.Client{
		Credential: auth.StaticCredential(registryWithAuthHost, auth.Credential{
			Username: username,
			Password: password,
		}),
	}, plainHTTP); err != nil {
		panic(err)
	}
}
