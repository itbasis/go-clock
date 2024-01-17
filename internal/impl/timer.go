package impl

import (
	"time"

	"github.com/itbasis/go-clock/v2/pkg"
)

// Timer represents a single event.
// The current time will be sent on C, unless the timer was created by AfterFunc.
type Timer struct {
	timer *time.Timer // realtime impl, if set
}

func NewTimer(d time.Duration) pkg.Timer {
	t := time.NewTimer(d)

	return &Timer{timer: t}
}

func NewTimerFunc(d time.Duration, f func()) pkg.Timer {
	t := time.AfterFunc(d, f)

	return &Timer{timer: t}
}

func (t *Timer) Chan() <-chan time.Time { return t.timer.C }

// Stop turns off the ticker.
func (t *Timer) Stop() bool { return t.timer.Stop() }

// Reset changes the expiry time of the timer
func (t *Timer) Reset(duration time.Duration) bool { return t.timer.Reset(duration) }
