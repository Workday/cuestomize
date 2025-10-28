// Package registryauth provides utilities for configuring authentication
// credentials for OCI registry clients, supporting both Kubernetes Secret
// and environment variable-based credential sources.
package registryauth

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	// UsernameEnvVar is the environment variable name for the username to use for registry authentication.
	UsernameEnvVar = "REGISTRY_USERNAME"
	// PasswordEnvVar is the environment variable name for the password to use for registry authentication.
	PasswordEnvVar = "REGISTRY_PASSWORD"
	// AccessTokenEnvVar is the environment variable name for the access token to use for registry authentication.
	AccessTokenEnvVar = "REGISTRY_ACCESS_TOKEN"
	// RefreshTokenEnvVar is the environment variable name for the refresh token to use for registry authentication.
	RefreshTokenEnvVar = "REGISTRY_REFRESH_TOKEN"
)

// ConfigureClient configures a remote client to fetch from the specified registry, with authentication if
// any is found.
func ConfigureClient(registry string, authSecret *corev1.Secret) (*auth.Client, error) {
	creds, err := configureAuth(authSecret)
	if err != nil {
		return nil, err
	}

	client := auth.DefaultClient
	if creds != nil {
		client.Credential = auth.StaticCredential(registry, *creds)
	}
	return client, nil
}

// configureAuth configures authentication based on the provided authSecret or environment variables.
// If no authentication is found, it returns nil, nil (no error).
func configureAuth(authSecret *corev1.Secret) (*auth.Credential, error) {
	if authSecret == nil {
		return getAuthFromEnv(), nil
	}

	creds := &auth.Credential{}

	for k, v := range authSecret.Data {
		switch k {
		case "username":
			creds.Username = string(v)
		case "password":
			creds.Password = string(v)
		case "accessToken":
			creds.AccessToken = string(v)
		case "refreshToken":
			creds.RefreshToken = string(v)
		}
	}

	return creds, nil
}

func getAuthFromEnv() *auth.Credential {
	username := os.Getenv(UsernameEnvVar)
	password := os.Getenv(PasswordEnvVar)
	accessToken := os.Getenv(AccessTokenEnvVar)
	refreshToken := os.Getenv(RefreshTokenEnvVar)

	if username == "" && password == "" && accessToken == "" && refreshToken == "" {
		return nil
	}

	return &auth.Credential{
		Username:     os.Getenv(UsernameEnvVar),
		Password:     os.Getenv(PasswordEnvVar),
		AccessToken:  os.Getenv(AccessTokenEnvVar),
		RefreshToken: os.Getenv(RefreshTokenEnvVar),
	}
}
