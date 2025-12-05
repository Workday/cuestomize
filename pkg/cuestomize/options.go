package cuestomize

import (
	"fmt"

	"github.com/Workday/cuestomize/pkg/cuestomize/model"
)

// CuestomizeOptions holds configuration options for the Cuestomize function.
type CuestomizeOptions struct {
	ModelProvider model.Provider
}

func (o *CuestomizeOptions) validate() error {
	if o.ModelProvider == nil {
		return fmt.Errorf("model provider is required")
	}
	return nil
}

// WithModelProvider sets the model provider to use for fetching the CUE model.
func WithModelProvider(provider model.Provider) func(*CuestomizeOptions) {
	return func(opts *CuestomizeOptions) {
		opts.ModelProvider = provider
	}
}
