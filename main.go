package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Workday/cuestomize/api"
	krm "github.com/Workday/cuestomize/internal/pkg/cuestomize"
	"github.com/Workday/cuestomize/internal/pkg/processor"
	"github.com/rs/zerolog"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

const (
	// Version is the version of the Cuestomize tool.
	Version = "0.0.1"

	// LogLevelEnvVar is the name of the environment variable that can be used to set the log level.
	LogLevelEnvVar = "LOG_LEVEL"
)

func main() {
	if err := setupLogging(); err != nil {
		os.Exit(1)
	}

	config := new(api.KRMInput)
	fn, err := krm.NewBuilder().SetConfig(config).Build()
	if err != nil {
		log.Fatalf("failed to build KRM function: %v", err)
	}

	p := processor.NewSimpleProcessor(config, kio.FilterFunc(fn), true)
	cmd := command.Build(p, command.StandaloneDisabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// setupLogging configures the global logging level based on the log level environment variable.
func setupLogging() error {
	logLevel := os.Getenv(LogLevelEnvVar)
	if logLevel == "" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level from environment variable %s: %w", LogLevelEnvVar, err)
	}
	zerolog.SetGlobalLevel(level)
	return nil
}
