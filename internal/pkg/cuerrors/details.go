package cuerrors

import (
	"fmt"

	"cuelang.org/go/cue/errors"
)

// ErrorWithDetails formats an error message with additional details from the provided error.
func ErrorWithDetails(err error, format string, args ...any) error {
	return ErrorWithDetailsCfg(err, nil, format, args...)
}

// ErrorWithDetailsCfg formats an error message with additional details from the provided error
// and uses the provided errors.Config to extract details.
func ErrorWithDetailsCfg(err error, cfg *errors.Config, format string, args ...any) error {
	errString := fmt.Sprintf(format, args...)

	details := errors.Details(err, cfg)

	return fmt.Errorf("%s: %s", errString, details)
}
