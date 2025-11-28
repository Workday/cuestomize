// Package cuerrors provides utilities for enhanced error formatting and detailed error reporting
// when dealing with CUE language errors. It includes functionality to format errors with
// additional context and details using CUE's built-in error configuration system.
//
// The main component is the Detailer type, which wraps CUE's errors.Config to provide
// consistent error formatting.
package cuerrors

import (
	"fmt"

	"cuelang.org/go/cue/errors"
)

// Detailer interface defines methods for formatting errors with additional CUE details.
type Detailer interface {
	// ErrorWithDetails formats an error message with additional details from the provided error.
	ErrorWithDetails(err error, format string, args ...any) error
}

// DefaultDetailer struct can be used to format CUE errors with additional details.
type DefaultDetailer struct {
	Cfg errors.Config
}

// NewDefaultDetailer creates a new DefaultDetailer with the specified current working directory (cwd).
// In formatted errors, file paths will be made relative to this directory.
func NewDefaultDetailer(cwd string) DefaultDetailer {
	return DefaultDetailer{
		Cfg: errors.Config{
			Cwd: cwd,
		},
	}
}

// ErrorWithDetails formats an error message with additional details from the provided error.
func (d DefaultDetailer) ErrorWithDetails(err error, format string, args ...any) error {
	errString := fmt.Sprintf(format, args...)

	details := errors.Details(err, &d.Cfg)

	return fmt.Errorf("%s: %s%w", errString, details, errWrapper{err: err})
}

// EmptyDetailer is a Detailer implementation that does not provide any additional details.
// Useful when you want the error without any CUE-specific formatting.
type EmptyDetailer struct{}

// ErrorWithDetails returns a formatted error message without any additional details.
func (e EmptyDetailer) ErrorWithDetails(err error, format string, args ...any) error {
	format += ": %w"
	args = append(args, err)
	return fmt.Errorf(format, args...)
}

// errWrapper is used to wrap the original error in the detailed answer without having it
// appearing in the formatted string. This still allows users to get the source error after
// formatting it with details, to preserve the errors chain.
type errWrapper struct {
	err error `json:"-"`
}

func (errWrapper) Error() string {
	return ""
}

func (e errWrapper) Unwrap() error {
	return e.err
}
