package cuerrors

import (
	"context"
	"fmt"
)

// contextKey is how we find Detailers in a context.Context.
type contextKey struct{}

// notFoundError exists to carry an IsNotFound method.
type notFoundError struct{}

func (notFoundError) Error() string {
	return "no cuerrors.Detailer was present"
}

func (notFoundError) IsNotFound() bool {
	return true
}

// NewContext returns a new Context, derived from ctx, which carries the
// provided Detailer.
func NewContext(ctx context.Context, detailer Detailer) context.Context {
	return context.WithValue(ctx, contextKey{}, detailer)
}

// FromContext returns a Detailer from ctx or an error if no Detailer is found.
func FromContext(ctx context.Context) (Detailer, error) {
	v := ctx.Value(contextKey{})
	if v == nil {
		return Detailer{}, notFoundError{}
	}

	switch v := v.(type) {
	case Detailer:
		return v, nil
	default:
		// Not reached.
		panic(fmt.Sprintf("unexpected value type for logr context key: %T", v))
	}
}

// FromContextOrDefault returns a Detailer from ctx.  If no Detailer is found, this
// returns a default Detailer.
func FromContextOrDefault(ctx context.Context) Detailer {
	if detailer, err := FromContext(ctx); err == nil {
		return detailer
	}
	return Detailer{}
}
