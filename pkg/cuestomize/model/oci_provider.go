package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/oci/fetcher"
	"github.com/go-logr/logr"
	"oras.land/oras-go/v2/registry/remote/auth"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// ociModelProviderOptions holds configuration options for OCIModelProvider.
type ociModelProviderOptions struct {
	Registry   string
	Repo       string
	Tag        string
	PlainHTTP  bool
	Client     *auth.Client
	WorkingDir string
}

// WithRemote configures the OCI remote to fetch the CUE model from.
func WithRemote(registry, repo, tag string) func(*ociModelProviderOptions) {
	return func(opts *ociModelProviderOptions) {
		opts.Registry = registry
		opts.Repo = repo
		opts.Tag = tag
	}
}

// WithPlainHTTP configures whether to use plain HTTP when fetching from the OCI registry.
func WithPlainHTTP(plainHTTP bool) func(*ociModelProviderOptions) {
	return func(opts *ociModelProviderOptions) {
		opts.PlainHTTP = plainHTTP
	}
}

// WithWorkingDir configures the working directory where the CUE model will be stored.
func WithWorkingDir(workingDir string) func(*ociModelProviderOptions) {
	return func(opts *ociModelProviderOptions) {
		opts.WorkingDir = workingDir
	}
}

// WithClient configures the OCI registry client to use when fetching the CUE model.
func WithClient(client *auth.Client) func(*ociModelProviderOptions) {
	return func(opts *ociModelProviderOptions) {
		opts.Client = client
	}
}

// OCIModelProvider is a model provider that fetches the CUE model from an OCI registry.
type OCIModelProvider struct {
	registry   string
	repo       string
	tag        string
	plainHTTP  bool
	workingDir string
	client     *auth.Client
}

// NewOCIModelProviderFromConfigAndItems creates a new OCIModelProvider based on the provided KRMInput configuration and input items.
func NewOCIModelProviderFromConfigAndItems(config *api.KRMInput, items []*kyaml.RNode) (*OCIModelProvider, error) {
	if config.RemoteModule == nil {
		return nil, fmt.Errorf("remote module configuration is missing")
	}
	client, err := config.GetRemoteClient(items)
	if err != nil {
		return nil, fmt.Errorf("failed to configure remote client: %w", err)
	}
	return New(
		WithRemote(config.RemoteModule.Registry, config.RemoteModule.Repo, config.RemoteModule.Tag),
		WithPlainHTTP(config.RemoteModule.PlainHTTP),
		WithClient(client),
	)
}

// New creates a new OCIModelProvider with the given options.
func New(opts ...func(*ociModelProviderOptions)) (*OCIModelProvider, error) {
	options := &ociModelProviderOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.Client == nil {
		options.Client = auth.DefaultClient
	}

	if options.WorkingDir == "" {
		workingdir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		options.WorkingDir = workingdir
	}

	return &OCIModelProvider{
		registry:   options.Registry,
		repo:       options.Repo,
		tag:        options.Tag,
		plainHTTP:  options.PlainHTTP,
		workingDir: options.WorkingDir,
		client:     options.Client,
	}, nil
}

// Path returns the local file system path to the CUE model.
func (p *OCIModelProvider) Path() string {
	return p.workingDir
}

// Get fetches the CUE model from the OCI registry and stores it in the working directory.
func (p *OCIModelProvider) Get(ctx context.Context) error {
	log := logr.FromContextOrDiscard(ctx).V(4).WithValues(
		"registry", p.registry, "repo", p.repo, "tag", p.tag, "workingDir", p.workingDir,
	)

	log.Info("fetching from OCI registry", "plainHTTP", p.plainHTTP)

	err := fetcher.FetchFromOCIRegistry(
		ctx,
		p.client,
		p.workingDir,
		p.registry,
		p.repo,
		p.tag,
		p.plainHTTP,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch from OCI registry: %w", err)
	}

	// best-effort validation of module structure
	_, err = os.Stat(filepath.Join(p.workingDir, "cue.mod"))
	if err != nil {
		log.V(-1).Info("cue.mod directory not found in artifact. This might cause Cuestomize issues interacting with the module.", "error", err)
	}

	return nil
}
