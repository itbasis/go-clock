package mock

import (
	"time"

	"github.com/itbasis/go-clock/v2/internal"
)

// Ticker holds a channel that receives "ticks" at regular intervals.
type Ticker struct {
	c       chan time.Time
	next    time.Time     // next tick time
	mock    *Mock         // mock clock, if set
	d       time.Duration // time between ticks
	stopped bool          // True if stopped, false if running
}

func NewTicker(c chan time.Time, m *Mock, duration time.Duration) *Ticker {
	return &Ticker{
		c:    c,
		mock: m,
		d:    duration,
		next: m.now.Add(duration),
	}
}

func (t *Ticker) Chan() <-chan time.Time { return t.c }

// Stop turns off the ticker.
func (t *Ticker) Stop() {
	t.mock.mu.Lock()
	t.mock.removeClockTimer(t)
	t.stopped = true
	t.mock.mu.Unlock()
}

// Reset resets the ticker to a new duration.
func (t *Ticker) Reset(duration time.Duration) {
	t.mock.mu.Lock()
	defer t.mock.mu.Unlock()

	if t.stopped {
		t.mock.timers = append(t.mock.timers, t)
		t.stopped = false
	}

	t.d = duration
	t.next = t.mock.now.Add(duration)
}

func (t *Ticker) Next() time.Time { return t.next }

func (t *Ticker) Tick(now time.Time) {
	select {
	case t.c <- now:
	default:
	}

	t.mock.mu.Lock()
	t.next = now.Add(t.d)
	t.mock.mu.Unlock()

	internal.Gosched()
}
