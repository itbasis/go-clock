clock
=====

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/mod/github.com/itbasis/go-clock)
![GitHub Release](https://img.shields.io/github/v/release/itbasis/go-clock)
[![codecov](https://codecov.io/gh/itbasis/go-clock/graph/badge.svg?token=NgqYRaNbNb)](https://codecov.io/gh/itbasis/go-clock)
[![Go Report Card](https://goreportcard.com/badge/github.com/itbasis/go-clock)](https://goreportcard.com/report/github.com/itbasis/go-clock)

> [!IMPORTANT]
> This repository will no longer be developed - please use [clockwork](https://github.com/jonboulle/clockwork), which has similar functionality but a larger community

Clock is a small library for mocking time in Go. It provides an interface
around the standard library's [`time`][time] package so that the application
can use the realtime clock while tests can use the mock clock.

This module is no longer maintained.

[time]: https://pkg.go.dev/github.com/itbasis/go-clock

## Usage

### Realtime Clock

Your application can maintain a `Clock` variable that will allow realtime and
mock clocks to be interchangeable. For example, if you had an `Application` type:

```go
import "github.com/itbasis/go-clock/v2"

type Application struct {
	Clock clock.Clock
}
```

You could initialize it to use the realtime clock like this:

```go
var app Application
app.Clock = clock.New()
...
```

Then all timers and time-related functionality should be performed from the
`Clock` variable.


### Mocking time

In your tests, you will want to use a `Mock` clock:

```go
import (
	"testing"

	"github.com/itbasis/go-clock/v2"
)

func TestApplication_DoSomething(t *testing.T) {
	mockClock := clock.NewMock()
	app := Application{Clock: mockClock}
	...
}
```

Now that you've initialized your application to use the mock clock, you can
adjust the time programmatically. The mock clock always starts from the Unix
epoch (midnight UTC on Jan 1, 1970).


### Controlling time

The mock clock provides the same functions that the standard library's `time`
package provides. For example, to find the current time, you use the `Now()`
function:

```go
mock := clock.NewMock()

// Find the current time.
mock.Now().UTC() // 1970-01-01 00:00:00 +0000 UTC

// Move the clock forward.
mock.Add(2 * time.Hour)

// Check the time again. It's 2 hours later!
mock.Now().UTC() // 1970-01-01 02:00:00 +0000 UTC
```

Timers and Tickers are also controlled by this same mock clock. They will only
execute when the clock is moved forward:

```go
mock := clock.NewMock()
count := 0

// Kick off a timer to increment every 1 mock second.
go func() {
    ticker := mock.Ticker(1 * time.Second)
    for {
        <-ticker.C
        count++
    }
}()
runtime.Gosched()

// Move the clock forward 10 seconds.
mock.Add(10 * time.Second)

// This prints 10.
fmt.Println(count)
```

### Working with context

It is possible to put clock into context without passing it directly to the function:

```go
clock := clock.New()
ctx := clock.WithContext(context.Background(), clock)
do(ctx)

func do(ctx context.Context) {
	clock := clock.FromContext(ctx)
	// do some staff..
}

```
