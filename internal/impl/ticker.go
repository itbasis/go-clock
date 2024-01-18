package impl

import (
	"time"

	"github.com/itbasis/go-clock/v2/pkg"
)

// Ticker holds a channel that receives "ticks" at regular intervals.
type Ticker struct {
	ticker *time.Ticker // realtime impl, if set
}

func NewTicker(d time.Duration) pkg.Ticker {
	t := time.NewTicker(d)

	return &Ticker{ticker: t}
}

func (t *Ticker) Chan() <-chan time.Time { return t.ticker.C }

// Stop turns off the ticker.
func (t *Ticker) Stop() { t.ticker.Stop() }

// Reset resets the ticker to a new duration.
func (t *Ticker) Reset(dur time.Duration) { t.ticker.Reset(dur) }
