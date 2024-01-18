package mock

import (
	"time"
)

// clockTimer represents an object with an associated start time.
type clockTicker interface {
	Next() time.Time
	Tick(time.Time)
}
