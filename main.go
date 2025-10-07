package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"golang.org/x/term"

	"github.com/Workday/cuestomize/api"
	krm "github.com/Workday/cuestomize/internal/pkg/cuestomize"
	"github.com/Workday/cuestomize/internal/pkg/processor"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

const (
	// LogLevelEnvVar is the name of the environment variable that can be used to set the log level.
	LogLevelEnvVar = "LOG_LEVEL"
)

// Version is the version of Cuestomize.
//
//go:embed semver
var Version string

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
	cmd := &cobra.Command{}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		cmd = command.Build(p, command.StandaloneEnabled, false)
	} else {
		cmd = command.Build(p, command.StandaloneDisabled, false)
	}
	cmd.Version = Version
	cmd.SetVersionTemplate("v{{.Version}}")

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
