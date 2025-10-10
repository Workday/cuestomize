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

// Detailer struct can be used to format CUE errors with additional details.
type Detailer struct {
	Cfg errors.Config
}

// NewDetailerWithCwd creates a new Detailer with the specified current working directory (cwd).
// In formatted errors, file paths will be made relative to this directory.
func NewDetailerWithCwd(cwd string) Detailer {
	return Detailer{
		Cfg: errors.Config{
			Cwd: cwd,
		},
	}
}

// ErrorWithDetails formats an error message with additional details from the provided error.
func (d *Detailer) ErrorWithDetails(err error, format string, args ...any) error {
	errString := fmt.Sprintf(format, args...)

	details := errors.Details(err, &d.Cfg)

	return fmt.Errorf("%s: %s", errString, details)
}
