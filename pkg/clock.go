package pkg

import (
	"context"
	"time"
)

// Re-export of time.Duration
type Duration = time.Duration

// Clock represents an interface to the functions in the standard library time
// package. Two implementations are available in the clock package. The first
// is a real-time clock which simply wraps the time package's functions. The
// second is a mock clock which will only change when
// programmatically adjusted.
type Clock interface { //nolint:interfacebloat
	After(d time.Duration) <-chan time.Time
	AfterFunc(d time.Duration, f func()) Timer
	Now() time.Time
	Since(t time.Time) time.Duration
	Until(t time.Time) time.Duration
	Sleep(d time.Duration)
	Tick(d time.Duration) <-chan time.Time
	Ticker(d time.Duration) Ticker
	Timer(d time.Duration) Timer
	WithDeadline(parent context.Context, d time.Time) (context.Context, context.CancelFunc)
	WithTimeout(parent context.Context, t time.Duration) (context.Context, context.CancelFunc)
}
