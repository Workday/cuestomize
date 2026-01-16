package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Workday/cuestomize/api"
	"github.com/Workday/cuestomize/pkg/oci/fetcher"
	"github.com/go-logr/logr"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/auth"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// OCIOption defines a functional option for configuring OCIModelProvider.
type OCIOption func(*ociModelProviderOptions)

// ociModelProviderOptions holds configuration options for OCIModelProvider.
type ociModelProviderOptions struct {
	Reference  registry.Reference
	PlainHTTP  bool
	Client     *auth.Client
	WorkingDir string
}

// WithRemoteParts configures the OCI remote to fetch the CUE model from an OCI registry.
//
// Panics if the provided registry, repo, and tag do not form a valid OCI reference.
func WithRemoteParts(reg, repo, tag string) OCIOption {
	reference := fmt.Sprintf("%s/%s", reg, repo)
	if tag != "" {
		reference = fmt.Sprintf("%s:%s", reference, tag)
	}
	ref, err := registry.ParseReference(reference)
	if err != nil {
		panic(fmt.Sprintf("invalid reference: %v", err))
	}
	return func(opts *ociModelProviderOptions) {
		opts.Reference = ref
	}
}

func WithRemote(ref registry.Reference) OCIOption {
	return func(opts *ociModelProviderOptions) {
		opts.Reference = ref
	}
}

// WithPlainHTTP configures whether to use plain HTTP when fetching from the OCI registry.
func WithPlainHTTP(plainHTTP bool) OCIOption {
	return func(opts *ociModelProviderOptions) {
		opts.PlainHTTP = plainHTTP
	}
}

// WithWorkingDir configures the working directory where the CUE model will be stored.
func WithWorkingDir(workingDir string) OCIOption {
	return func(opts *ociModelProviderOptions) {
		opts.WorkingDir = workingDir
	}
}

// WithClient configures the OCI registry client to use when fetching the CUE model.
func WithClient(client *auth.Client) OCIOption {
	return func(opts *ociModelProviderOptions) {
		opts.Client = client
	}
}

// OCIModelProvider is a model provider that fetches the CUE model from an OCI registry.
type OCIModelProvider struct {
	reference  registry.Reference
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

	reference, err := config.RemoteModule.GetReference()
	if err != nil {
		return nil, fmt.Errorf("failed to get reference: %w", err)
	}

	return New(
		WithRemote(reference),
		WithPlainHTTP(config.RemoteModule.PlainHTTP),
		WithClient(client),
	)
}

// New creates a new OCIModelProvider with the given options.
func New(opts ...OCIOption) (*OCIModelProvider, error) {
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
		reference:  options.Reference,
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
		"registry", p.reference.Registry, "repo", p.reference.Repository, "tag", p.reference.Reference, "workingDir", p.workingDir,
	)

	log.Info("fetching from OCI registry", "plainHTTP", p.plainHTTP)

	err := fetcher.FetchFromOCIRegistry(
		ctx,
		p.client,
		p.workingDir,
		p.reference,
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
