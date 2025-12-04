package cuestomize

import (
	"fmt"

	"github.com/Workday/cuestomize/pkg/cuestomize/model"
)

// CustomizeOptions holds configuration options for the Cuestomize function.
type CustomizeOptions struct {
	ModelProvider model.Provider
}

func (o *CustomizeOptions) validate() error {
	if o.ModelProvider == nil {
		return fmt.Errorf("model provider is required")
	}
	return nil
}

// WithModelProvider sets the model provider to use for fetching the CUE model.
func WithModelProvider(provider model.Provider) func(*CustomizeOptions) {
	return func(opts *CustomizeOptions) {
		opts.ModelProvider = provider
	}
}
