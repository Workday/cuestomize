package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/Workday/cuestomize/api"
	krm "github.com/Workday/cuestomize/internal/pkg/cuestomize"
	"github.com/Workday/cuestomize/internal/pkg/processor"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"

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
	ctx := context.Background()
	if err := setupLogging(ctx); err != nil {
		os.Exit(1)
	}

	config := new(api.KRMInput)
	fn, err := krm.NewBuilder().SetConfig(config).Build()
	if err != nil {
		log.Fatalf("failed to build KRM function: %v", err)
	}

	p := processor.NewSimpleProcessor(config, kio.FilterFunc(fn), true)
	cmd := command.Build(p, command.StandaloneDisabled, false)
	cmd.Version = Version
	cmd.SetVersionTemplate("v{{.Version}}")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// setupLogging configures the global logging level based on the log level environment variable.
func setupLogging(ctx context.Context) (context.Context, error) {
	logLevel := os.Getenv(LogLevelEnvVar)
	if logLevel == "" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level from environment variable %s: %w", LogLevelEnvVar, err)
	}
	zerolog.SetGlobalLevel(level)

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
