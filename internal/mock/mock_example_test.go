package mock_test

import (
	"fmt"
	"time"

	"github.com/itbasis/go-clock/v2/internal"
	"github.com/itbasis/go-clock/v2/internal/mock"
)

func ExampleMock_After() {
	var ( // Create a new mock clock.
		clock = mock.NewMock()
		count counter
	)

	ready := make(chan struct{})
	// Create a channel to execute after 10 mock seconds.
	go func() {
		ch := clock.After(10 * time.Second)

		close(ready)

		<-ch
		count.incr()
	}()
	<-ready

	// Print the starting value.
	fmt.Printf("%s: %d\n", clock.Now().UTC(), count.get())

	// Move the clock forward 5 seconds and print the value again.
	clock.Add(5 * time.Second)
	fmt.Printf("%s: %d\n", clock.Now().UTC(), count.get())

	// Move the clock forward 5 seconds to the tick time and check the value.
	clock.Add(5 * time.Second)
	fmt.Printf("%s: %d\n", clock.Now().UTC(), count.get())

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: 0
	// 1970-01-01 00:00:05 +0000 UTC: 0
	// 1970-01-01 00:00:10 +0000 UTC: 1
}

func ExampleMock_AfterFunc() {
	var (
		// Create a new mock clock.
		clock = mock.NewMock()
		count counter
	)

	count.incr()

	// Execute a function after 10 mock seconds.
	clock.AfterFunc(
		10*time.Second, func() {
			count.incr()
		},
	)
	internal.Gosched()

	// Print the starting value.
	fmt.Printf("%s: %d\n", clock.Now().UTC(), count.get())

	// Move the clock forward 10 seconds and print the new value.
	clock.Add(10 * time.Second)
	fmt.Printf("%s: %d\n", clock.Now().UTC(), count.get())

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: 1
	// 1970-01-01 00:00:10 +0000 UTC: 2
}

func ExampleMock_Sleep() {
	var (
		// Create a new mock clock.
		clock = mock.NewMock()
		count counter
	)

	// Execute a function after 10 mock seconds.
	go func() {
		clock.Sleep(10 * time.Second)
		count.incr()
	}()

	internal.Gosched()

	// Print the starting value.
	fmt.Printf("%s: %d\n", clock.Now().UTC(), count.get())

	// Move the clock forward 10 seconds and print the new value.
	clock.Add(10 * time.Second)
	fmt.Printf("%s: %d\n", clock.Now().UTC(), count.get())

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: 0
	// 1970-01-01 00:00:10 +0000 UTC: 1
}

func ExampleMock_Ticker() {
	var (
		// Create a new mock clock.
		clock = mock.NewMock()
		count counter
	)

	ready := make(chan struct{})
	// Increment count every mock second.
	go func() {
		ticker := clock.Ticker(1 * time.Second)

		close(ready)

		for {
			<-ticker.Chan()
			count.incr()
		}
	}()
	<-ready

	// Move the clock forward 10 seconds and print the new value.
	clock.Add(10 * time.Second)
	fmt.Printf("Count is %d after 10 seconds\n", count.get())

	// Move the clock forward 5 more seconds and print the new value.
	clock.Add(5 * time.Second)
	fmt.Printf("Count is %d after 15 seconds\n", count.get())

	// Output:
	// Count is 10 after 10 seconds
	// Count is 15 after 15 seconds
}

func ExampleMock_Timer() {
	var ( // Create a new mock clock.
		clock = mock.NewMock()
		count counter
	)

	ready := make(chan struct{})
	// Increment count after a mock second.
	go func() {
		timer := clock.Timer(1 * time.Second)

		close(ready)

		<-timer.Chan()
		count.incr()
	}()
	<-ready

	// Move the clock forward 10 seconds and print the new value.
	clock.Add(10 * time.Second)
	fmt.Printf("Count is %d after 10 seconds\n", count.get())

	// Output:
	// Count is 1 after 10 seconds
}
