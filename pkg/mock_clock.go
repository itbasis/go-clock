package pkg

import "time"

type Mock interface {
	Clock

	Add(duration time.Duration)
	Set(time time.Time)

	WaitForAllTimers() time.Time
}
