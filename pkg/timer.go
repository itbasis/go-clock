package pkg

import "time"

type Timer interface {
	Chan() <-chan time.Time

	Stop() bool
	Reset(duration time.Duration) bool
}
