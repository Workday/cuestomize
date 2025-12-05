package cuestomize

import (
	"fmt"

	"github.com/Workday/cuestomize/pkg/cuestomize/model"
)

// Option defines a functional option for configuring Cuestomize.
type Option func(*options)

// options holds configuration options for the Cuestomize function.
type options struct {
	ModelProvider model.Provider
}

func (o *options) validate() error {
	if o.ModelProvider == nil {
		return fmt.Errorf("model provider is required")
	}
	return nil
}

// WithModelProvider sets the model provider to use for fetching the CUE model.
func WithModelProvider(provider model.Provider) Option {
	return func(opts *options) {
		opts.ModelProvider = provider
	}
}
