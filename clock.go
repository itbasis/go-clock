package clock

import (
	"github.com/itbasis/go-clock/v2/internal/impl"
	"github.com/itbasis/go-clock/v2/internal/mock"
	"github.com/itbasis/go-clock/v2/pkg"
)

// New returns an instance of a real-time clock.
func New() pkg.Clock {
	return impl.NewClock()
}

// NewMock returns an instance of a mock clock.
// The current time of the mock clock on initialization is the Unix epoch.
func NewMock() pkg.Mock {
	return mock.NewMock()
}
