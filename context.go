package clock

import (
	"context"

	"github.com/itbasis/go-clock/v2/pkg"
)

// Used as a context key which holds clock value
type ctxClock struct{}

// WithContext creates child context with embedded clock implementation
func WithContext(ctx context.Context, clock pkg.Clock) context.Context {
	return context.WithValue(ctx, ctxClock{}, clock)
}

// FromContext returns the implementation of clock associated with provided context.
// It returns new instance of real clock implementation if not present.
func FromContext(ctx context.Context) pkg.Clock {
	if clock, ok := ctx.Value(ctxClock{}).(pkg.Clock); ok {
		return clock
	}

	return New()
}
