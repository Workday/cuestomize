package cuerrors

import (
	"fmt"

	"cuelang.org/go/cue/errors"
)

type Detailer struct {
	Cfg errors.Config
}

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
