package cmd

import (
	"fmt"
	"os"

	"github.com/Workday/cuestomize/api"
	krm "github.com/Workday/cuestomize/internal/pkg/cuestomize"
	"github.com/Workday/cuestomize/internal/pkg/processor"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/yaml"
)

func NewCLICommand() *cobra.Command {
	var configPath string

	var validateOnly bool
	var module string
	var includeAll bool
	var plainHTTP bool

	cliCmd := &cobra.Command{
		Use:   "cli",
		Short: "Run Cuestomize as a CLI tool",
		Long:  "Run Cuestomize as a standalone CLI tool, reading resources from stdin and configuration from a file.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if configPath != "" && !(validateOnly || module != "" || includeAll) {
				return fmt.Errorf("--config cannot be used with --validate-only, --module, or --include-all")
			}
			if configPath == "" && module == "" {
				return fmt.Errorf("either --config or --module must be specified")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			config := new(api.KRMInput)
			builder := krm.NewBuilder().SetConfig(config)

			if configPath != "" {
				b, err := os.ReadFile(configPath)
				if err != nil {
					return fmt.Errorf("failed to read config file: %w", err)
				}
				if err := yaml.UnmarshalStrict(b, config); err != nil {
					return fmt.Errorf("failed to unmarshal config: %w", err)
				}
			} else {
				info, err := os.Stat(module)
				switch err {
				case nil:
					if info.IsDir() {
						builder.SetResourcesPath(module)
					} else {
						return fmt.Errorf("error interpreting local module. Expected dir, got: %s", info.Mode().Type().String())
					}
				case os.ErrNotExist:
					// try to use module as remote reference
					// parse module to extract registry, repo, tag
					config.RemoteModule.Registry = module
					config.RemoteModule.Repo = module
					config.RemoteModule.Tag = "latest"
					config.RemoteModule.PlainHTTP = plainHTTP

					if validateOnly {
						config.Annotations = map[string]string{"config.cuestomize.io/validator": "true"}
					}

					if includeAll {
						config.Includes = []types.Selector{
							{ResId: resid.ResId{Gvk: resid.Gvk{Group: ".*", Version: ".*", Kind: ".*"}}},
						}
					}

					tempDir, err := os.MkdirTemp("", "cuestomize-cli-")
					if err != nil {
						return fmt.Errorf("failed to create temp dir: %w", err)
					}
					defer os.RemoveAll(tempDir)
					builder.SetResourcesPath(tempDir)
				default:
					return fmt.Errorf("failed to stat module: %w", err)
				}
			}

			fn, err := builder.Build()
			if err != nil {
				return fmt.Errorf("failed to build KRM function: %w", err)
			}
			p := processor.NewSimpleProcessor(config, kio.FilterFunc(fn), true)

			rw := &kio.ByteReadWriter{
				Reader:                cmd.InOrStdin(),
				Writer:                cmd.OutOrStdout(),
				OmitReaderAnnotations: true,
			}
			return framework.Execute(p, rw)
		},
	}
	cliCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to the Cuestomize configuration file")
	cliCmd.Flags().BoolVarP(&validateOnly, "validate-only", "v", false, "run in validate-only mode")
	cliCmd.Flags().StringVarP(&module, "module", "m", "", "the CUE module. It can be either a local path or an OCI repo reference")
	cliCmd.Flags().BoolVarP(&includeAll, "include-all", "a", false, "whether to forward all input resources to the CUE module as includes")
	cliCmd.Flags().BoolVar(&plainHTTP, "plain-http", false, "")
	return cliCmd
}
