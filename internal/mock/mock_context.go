package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/itbasis/go-clock/v2/pkg"
)

// propagateCancel arranges for child to be canceled when parent is.
func propagateCancel(parent context.Context, child *timerCtx) {
	if parent.Done() == nil {
		return // parent is never canceled
	}

	go func() {
		select {
		case <-parent.Done():
			child.cancel(parent.Err())
		case <-child.Done():
		}
	}()
}

type timerCtx struct {
	sync.Mutex

	clock    pkg.Clock
	parent   context.Context //nolint:containedctx
	deadline time.Time
	done     chan struct{}

	err   error
	timer pkg.Timer
}

func (c *timerCtx) cancel(err error) {
	c.Lock()

	defer c.Unlock()

	if c.err != nil {
		return // already canceled
	}

	c.err = err
	close(c.done)

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}

func (c *timerCtx) Deadline() (deadline time.Time, ok bool) { return c.deadline, true }

func (c *timerCtx) Done() <-chan struct{} { return c.done }

func (c *timerCtx) Err() error { return c.err }

func (c *timerCtx) Value(key interface{}) interface{} { return c.parent.Value(key) }

func (c *timerCtx) String() string {
	return fmt.Sprintf("clock.WithDeadline(%s [%s])", c.deadline, c.deadline.Sub(c.clock.Now()))
}
