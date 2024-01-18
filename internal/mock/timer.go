package mock

import (
	"time"

	"github.com/itbasis/go-clock/v2/internal"
)

// Timer represents a single event.
// The current time will be sent on C, unless the timer was created by AfterFunc.
type Timer struct {
	c       chan time.Time
	next    time.Time // next tick time
	mock    *Mock     // mock clock, if set
	fn      func()    // AfterFunc function, if set
	stopped bool      // True if stopped, false if running
}

func NewTimer(c chan time.Time, f func(), m *Mock, d time.Duration) *Timer {
	return &Timer{
		c:    c,
		fn:   f,
		mock: m,
		next: m.now.Add(d),
	}
}

func (t *Timer) Chan() <-chan time.Time { return t.c }

// Stop turns off the ticker.
func (t *Timer) Stop() bool {
	t.mock.mu.Lock()
	registered := !t.stopped
	t.mock.removeClockTimer(t)
	t.stopped = true
	t.mock.mu.Unlock()

	return registered
}

// Reset changes the expiry time of the timer
func (t *Timer) Reset(duration time.Duration) bool {
	t.mock.mu.Lock()
	t.next = t.mock.now.Add(duration)
	defer t.mock.mu.Unlock()

	registered := !t.stopped

	if t.stopped {
		t.mock.timers = append(t.mock.timers, t)
	}

	t.stopped = false

	return registered
}

func (t *Timer) Next() time.Time { return t.next }

func (t *Timer) Tick(now time.Time) {
	// a gosched() after ticking, to allow any consequences of the
	// tick to complete
	defer internal.Gosched()

	t.mock.mu.Lock()

	if t.fn != nil {
		// defer function execution until the lock is released, and
		defer func() { go t.fn() }()
	} else {
		t.c <- now
	}

	t.mock.removeClockTimer(t)
	t.stopped = true
	t.mock.mu.Unlock()
}
