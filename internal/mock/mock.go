package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/itbasis/go-clock/v2/internal"
	"github.com/itbasis/go-clock/v2/pkg"
)

// Mock represents a mock clock that only moves forward programmically.
// It can be preferable to a real-time clock when testing time-based functionality.
type Mock struct {
	// mu protects all other fields in this struct, and the data that they
	// point to.
	mu sync.Mutex

	now    time.Time    // current time
	timers clockTickers // tickers & timers
}

// NewMock returns an instance of a mock clock.
// The current time of the mock clock on initialization is the Unix epoch.
func NewMock() *Mock {
	return &Mock{now: time.Unix(0, 0)}
}

// After waits for the duration to elapse and then sends the current time on the returned channel.
func (m *Mock) After(d time.Duration) <-chan time.Time {
	return m.Timer(d).Chan()
}

// AfterFunc waits for the duration to elapse and then executes a function in its own goroutine.
// A Timer is returned that can be stopped.
func (m *Mock) AfterFunc(duration time.Duration, f func()) pkg.Timer {
	m.mu.Lock()

	defer m.mu.Unlock()

	ch := make(chan time.Time, 1)

	timer := NewTimer(ch, f, m, duration)

	m.timers = append(m.timers, timer)

	return timer
}

// Now returns the current wall time on the mock clock.
func (m *Mock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.now
}

// Since returns time since `t` using the mock clock's wall time.
func (m *Mock) Since(t time.Time) time.Duration {
	return m.Now().Sub(t)
}

// Until returns time until `t` using the mock clock's wall time.
func (m *Mock) Until(t time.Time) time.Duration {
	return t.Sub(m.Now())
}

// Sleep pauses the goroutine for the given duration on the mock clock.
// The clock must be moved forward in a separate goroutine.
func (m *Mock) Sleep(d time.Duration) {
	<-m.After(d)
}

// Tick is a convenience function for Ticker().
// It will return a ticker channel that cannot be stopped.
func (m *Mock) Tick(d time.Duration) <-chan time.Time { return m.Ticker(d).Chan() }

// Ticker creates a new instance of Ticker.
func (m *Mock) Ticker(duration time.Duration) pkg.Ticker {
	m.mu.Lock()

	defer m.mu.Unlock()

	ch := make(chan time.Time, 1)

	ticker := NewTicker(ch, m, duration)

	m.timers = append(m.timers, ticker)

	return ticker
}

// Timer creates a new instance of Timer.
func (m *Mock) Timer(duration time.Duration) pkg.Timer {
	m.mu.Lock()
	ch := make(chan time.Time, 1)

	timer := NewTimer(ch, nil, m, duration)

	m.timers = append(m.timers, timer)
	now := m.now
	m.mu.Unlock()
	m.runNextTimer(now)

	return timer
}

func (m *Mock) WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return m.WithDeadline(parent, m.Now().Add(timeout))
}

func (m *Mock) WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	if cur, ok := parent.Deadline(); ok && cur.Before(deadline) {
		// The current deadline is already sooner than the new one.
		return context.WithCancel(parent)
	}

	ctx := &timerCtx{clock: m, parent: parent, deadline: deadline, done: make(chan struct{})}
	propagateCancel(parent, ctx)

	dur := m.Until(deadline)

	if dur <= 0 {
		ctx.cancel(context.DeadlineExceeded) // deadline has already passed

		return ctx, func() {}
	}

	ctx.Lock()
	defer ctx.Unlock()

	if ctx.err == nil {
		ctx.timer = m.AfterFunc(
			dur, func() {
				ctx.cancel(context.DeadlineExceeded)
			},
		)
	}

	return ctx, func() { ctx.cancel(context.Canceled) }
}

// Add moves the current time of the mock clock forward by the specified duration.
// This should only be called from a single goroutine at a time.
func (m *Mock) Add(duration time.Duration) {
	// Calculate the final current time.
	m.mu.Lock()
	t := m.now.Add(duration)
	m.mu.Unlock()

	// Continue to execute timers until there are no more before the new time.
	for {
		if !m.runNextTimer(t) {
			break
		}
	}

	// Ensure that we end with the new time.
	m.mu.Lock()
	m.now = t
	m.mu.Unlock()

	// Give a small buffer to make sure that other goroutines get handled.
	internal.Gosched()
}

// Set sets the current time of the mock clock to a specific one.
// This should only be called from a single goroutine at a time.
func (m *Mock) Set(time time.Time) {
	// Continue to execute timers until there are no more before the new time.
	for {
		if !m.runNextTimer(time) {
			break
		}
	}

	// Ensure that we end with the new time.
	m.mu.Lock()
	m.now = time
	m.mu.Unlock()

	// Give a small buffer to make sure that other goroutines get handled.
	internal.Gosched()
}

// WaitForAllTimers sets the clock until all timers are expired
func (m *Mock) WaitForAllTimers() time.Time {
	// Continue to execute timers until there are no more
	for {
		m.mu.Lock()
		if len(m.timers) == 0 {
			m.mu.Unlock()

			return m.Now()
		}

		sort.Sort(m.timers)
		next := m.timers[len(m.timers)-1].Next()
		m.mu.Unlock()
		m.Set(next)
	}
}

// runNextTimer executes the next timer in chronological order and moves the
// current time to the timer's next tick time. The next time is not executed if
// its next time is after the max time. Returns true if a timer was executed.
func (m *Mock) runNextTimer(max time.Time) bool {
	m.mu.Lock()

	// Sort timers by time.
	sort.Sort(m.timers)

	// If we have no more timers then exit.
	if len(m.timers) == 0 {
		m.mu.Unlock()

		return false
	}

	// Retrieve next timer. Exit if next tick is after new time.
	t := m.timers[0]
	if t.Next().After(max) {
		m.mu.Unlock()

		return false
	}

	// Move "now" forward and unlock clock.
	m.now = t.Next()
	now := m.now
	m.mu.Unlock()

	// Execute timer.
	t.Tick(now)

	return true
}

// removeClockTimer removes a timer from m.timers. m.mu MUST be held
// when this method is called.
func (m *Mock) removeClockTimer(t clockTicker) {
	for i, timer := range m.timers {
		if timer == t {
			copy(m.timers[i:], m.timers[i+1:])
			m.timers[len(m.timers)-1] = nil
			m.timers = m.timers[:len(m.timers)-1]

			break
		}
	}

	sort.Sort(m.timers)
}
