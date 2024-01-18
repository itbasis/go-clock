package impl

import (
	"context"
	"time"

	"github.com/itbasis/go-clock/v2/pkg"
)

// Clock implements a real-time clock by simply wrapping the time package functions.
type Clock struct{}

func NewClock() pkg.Clock {
	return &Clock{}
}

func (c *Clock) After(d time.Duration) <-chan time.Time { return time.After(d) }

func (c *Clock) AfterFunc(d time.Duration, f func()) pkg.Timer { return NewTimerFunc(d, f) }

func (c *Clock) Now() time.Time { return time.Now() }

func (c *Clock) Since(t time.Time) time.Duration { return time.Since(t) }

func (c *Clock) Until(t time.Time) time.Duration { return time.Until(t) }

func (c *Clock) Sleep(d time.Duration) { time.Sleep(d) }

func (c *Clock) Tick(d time.Duration) <-chan time.Time {
	return c.Ticker(d).Chan()
}

func (c *Clock) Ticker(d time.Duration) pkg.Ticker { return NewTicker(d) }

func (c *Clock) Timer(d time.Duration) pkg.Timer { return NewTimer(d) }

func (c *Clock) WithDeadline(parent context.Context, d time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, d)
}

func (c *Clock) WithTimeout(parent context.Context, t time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, t)
}
