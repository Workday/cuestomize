package main

import (
	"context"
	"fmt"
	"os"

	"dagger/cuestomize/shared"

	"github.com/Workday/cuestomize/pkg/oci"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// main uploads e2e/testdata/cue module to local registry with and without authentication
// for e2e to run.
func main() {
	ctx := context.Background()

	registryNoAuthHost := os.Getenv(shared.RegistryHostVarName)
	if registryNoAuthHost == "" {
		panic(fmt.Sprintf("Environment variable %s is not set", shared.RegistryHostVarName))
	}
	registryWithAuthHost := os.Getenv(shared.RegistryAuthHostVarName)
	if registryWithAuthHost == "" {
		panic(fmt.Sprintf("Environment variable %s is not set", shared.RegistryAuthHostVarName))
	}

	username := os.Getenv(shared.RegistryUsernameVarName)
	password := os.Getenv(shared.RegistryPasswordVarName)

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
