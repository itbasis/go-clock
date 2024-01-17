package pkg

import "time"

type Ticker interface {
	Chan() <-chan time.Time

	Stop()
	Reset(duration time.Duration)
}
