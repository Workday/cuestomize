package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"

	"golang.org/x/term"

	"github.com/Workday/cuestomize/api"
	krm "github.com/Workday/cuestomize/internal/pkg/cuestomize"
	"github.com/Workday/cuestomize/internal/pkg/processor"
	"github.com/spf13/cobra"
	"github.com/go-logr/logr"

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
	ctx, err := setupLogging(context.Background())
	if err != nil {
		log.Fatalf("failed to set up logging: %v", err)
	}

	config := new(api.KRMInput)
	fn, err := krm.NewBuilder().SetConfig(config).Build(ctx)
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
	cmd.SetVersionTemplate("v{{.Version}}\n")

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

// setupLogging configures the global logging level based on the log level environment variable.
func setupLogging(ctx context.Context) (context.Context, error) {
	logLevel := os.Getenv(LogLevelEnvVar)

	lvl := slog.LevelInfo
	if logLevel != "" {
		err := lvl.UnmarshalText([]byte(logLevel))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal log level from environment variable %s: %w", LogLevelEnvVar, err)
		}
	}

	log := logr.FromSlogHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))

	return logr.NewContext(ctx, log), nil
}
