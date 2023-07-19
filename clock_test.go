package clock_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/itbasis/go-clock"
	"github.com/itbasis/go-clock/internal"
)

// counter is an atomic uint32 that can be incremented easily.  It's
// useful for asserting things have happened in tests.
type counter struct {
	count uint32
}

func (c *counter) incr() {
	atomic.AddUint32(&c.count, 1)
}

func (c *counter) get() uint32 {
	return atomic.LoadUint32(&c.count)
}

// Ensure that the clock's After channel sends at the correct time.
func TestClock_After(t *testing.T) {
	start := time.Now()
	<-clock.New().After(20 * time.Millisecond)
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock's AfterFunc executes at the correct time.
func TestClock_AfterFunc(t *testing.T) {
	var (
		ok bool
		wg sync.WaitGroup
	)

	wg.Add(1)

	start := time.Now()

	clock.New().AfterFunc(20*time.Millisecond, func() {
		ok = true
		wg.Done()
	})
	wg.Wait()

	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}

	if !ok {
		t.Fatal("Function did not run")
	}
}

// Ensure that the clock's time matches the standary library.
func TestClock_Now(t *testing.T) {
	a := time.Now().Round(time.Second)
	b := clock.New().Now().Round(time.Second)

	if !a.Equal(b) {
		t.Errorf("not equal: %s != %s", a, b)
	}
}

// Ensure that the clock sleeps for the appropriate amount of time.
func TestClock_Sleep(t *testing.T) {
	start := time.Now()
	clock.New().Sleep(20 * time.Millisecond)
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock ticks correctly.
func TestClock_Tick(t *testing.T) {
	start := time.Now()
	c := clock.New().Tick(20 * time.Millisecond)
	<-c
	<-c

	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 50*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock's ticker ticks correctly.
func TestClock_Ticker(t *testing.T) {
	start := time.Now()
	ticker := clock.New().Ticker(50 * time.Millisecond)
	<-ticker.C
	<-ticker.C

	dur := time.Since(start)

	if dur < 100*time.Millisecond || dur > 200*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock's ticker can stop correctly.
func TestClock_Ticker_Stp(t *testing.T) {
	ticker := clock.New().Ticker(20 * time.Millisecond)
	<-ticker.C
	ticker.Stop()
	select {
	case <-ticker.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's ticker can reset correctly.
func TestClock_Ticker_Rst(t *testing.T) {
	start := time.Now()
	ticker := clock.New().Ticker(20 * time.Millisecond)
	<-ticker.C
	ticker.Reset(5 * time.Millisecond)
	<-ticker.C

	dur := time.Since(start)

	if dur >= 30*time.Millisecond {
		t.Fatal("took more than 30ms")
	}

	ticker.Stop()
}

// Ensure that the clock's ticker can stop and then be reset correctly.
func TestClock_Ticker_Stop_Rst(t *testing.T) {
	start := time.Now()
	ticker := clock.New().Ticker(20 * time.Millisecond)
	<-ticker.C
	ticker.Stop()

	select {
	case <-ticker.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
	ticker.Reset(5 * time.Millisecond)
	<-ticker.C

	dur := time.Since(start)

	if dur >= 60*time.Millisecond {
		t.Fatal("took more than 60ms")
	}

	ticker.Stop()
}

// Ensure that the clock's timer waits correctly.
func TestClock_Timer(t *testing.T) {
	start := time.Now()
	timer := clock.New().Timer(20 * time.Millisecond)
	<-timer.C

	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}

	if timer.Stop() {
		t.Fatal("timer still running")
	}
}

// Ensure that the clock's timer can be stopped.
func TestClock_Timer_Stop(t *testing.T) {
	timer := clock.New().Timer(20 * time.Millisecond)

	if !timer.Stop() {
		t.Fatal("timer not running")
	}

	if timer.Stop() {
		t.Fatal("timer wasn't cancelled")
	}

	select {
	case <-timer.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's timer can be reset.
func TestClock_Timer_Reset(t *testing.T) {
	start := time.Now()
	timer := clock.New().Timer(10 * time.Millisecond)

	if !timer.Reset(20 * time.Millisecond) {
		t.Fatal("timer not running")
	}

	<-timer.C

	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

func TestClock_NegativeDuration(t *testing.T) {
	clock := clock.NewMock()
	timer := clock.Timer(-time.Second)
	select {
	case <-timer.C:
	default:
		t.Fatal("timer should have fired immediately")
	}
}

// Ensure reset can be called immediately after reading channel
func TestClock_Timer_Reset_Unlock(t *testing.T) {
	var (
		clock = clock.NewMock()
		timer = clock.Timer(1 * time.Second)
		wg    sync.WaitGroup
	)

	wg.Add(1)

	go func() {
		defer wg.Done()

		select { //nolint:gosimple
		case <-timer.C:
			timer.Reset(1 * time.Second)
		}

		select { //nolint:gosimple
		case <-timer.C:
		}
	}()

	clock.Add(2 * time.Second)
	wg.Wait()
}

// Ensure that the mock's After channel sends at the correct time.
func TestMock_After(t *testing.T) {
	var (
		ok    int32
		clock = clock.NewMock()
	)

	// Create a channel to execute after 10 mock seconds.
	ch := clock.After(10 * time.Second)

	go func(ch <-chan time.Time) {
		<-ch
		atomic.StoreInt32(&ok, 1)
	}(ch)

	// Move clock forward to just before the time.
	clock.Add(9 * time.Second)

	if atomic.LoadInt32(&ok) == 1 {
		t.Fatal("too early")
	}

	// Move clock forward to the after channel's time.
	clock.Add(1 * time.Second)

	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too late")
	}
}

// Ensure that the mock's After channel doesn't block on write.
func TestMock_UnusedAfter(t *testing.T) {
	mock := clock.NewMock()
	mock.After(1 * time.Millisecond)

	done := make(chan bool, 1)
	go func() {
		mock.Add(1 * time.Second)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("mock.Add hung")
	}
}

// Ensure that the mock's AfterFunc executes at the correct time.
func TestMock_AfterFunc(t *testing.T) {
	var (
		ok    int32
		clock = clock.NewMock()
	)

	// Execute function after duration.
	clock.AfterFunc(10*time.Second, func() {
		atomic.StoreInt32(&ok, 1)
	})

	// Move clock forward to just before the time.
	clock.Add(9 * time.Second)

	if atomic.LoadInt32(&ok) == 1 {
		t.Fatal("too early")
	}

	// Move clock forward to the after channel's time.
	clock.Add(1 * time.Second)

	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too late")
	}
}

// Ensure that the mock's AfterFunc doesn't execute if stopped.
func TestMock_AfterFunc_Stop(t *testing.T) {
	// Execute function after duration.
	clock := clock.NewMock()
	timer := clock.AfterFunc(10*time.Second, func() {
		t.Fatal("unexpected function execution")
	})

	internal.Gosched()

	// Stop timer & move clock forward.
	timer.Stop()
	clock.Add(10 * time.Second)
	internal.Gosched()
}

// Ensure that the mock's current time can be changed.
func TestMock_Now(t *testing.T) {
	clock := clock.NewMock()

	if now := clock.Now(); !now.Equal(time.Unix(0, 0)) {
		t.Fatalf("expected epoch, got: %v", now)
	}

	// Add 10 seconds and check the time.
	clock.Add(10 * time.Second)

	if now := clock.Now(); !now.Equal(time.Unix(10, 0)) {
		t.Fatalf("expected epoch, got: %v", now)
	}
}

func TestMock_Since(t *testing.T) {
	clock := clock.NewMock()

	beginning := clock.Now()
	clock.Add(500 * time.Second)

	if since := clock.Since(beginning); since.Seconds() != 500 {
		t.Fatalf("expected 500 since beginning, actually: %v", since.Seconds())
	}
}

func TestMock_Until(t *testing.T) {
	clock := clock.NewMock()

	end := clock.Now().Add(500 * time.Second)
	if dur := clock.Until(end); dur.Seconds() != 500 {
		t.Fatalf("expected 500s duration between `clock` and `end`, actually: %v", dur.Seconds())
	}

	clock.Add(100 * time.Second)

	if dur := clock.Until(end); dur.Seconds() != 400 {
		t.Fatalf("expected 400s duration between `clock` and `end`, actually: %v", dur.Seconds())
	}
}

// Ensure that the mock can sleep for the correct time.
func TestMock_Sleep(t *testing.T) {
	var (
		ok    int32
		clock = clock.NewMock()
	)

	// Create a channel to execute after 10 mock seconds.
	go func() {
		clock.Sleep(10 * time.Second)
		atomic.StoreInt32(&ok, 1)
	}()

	internal.Gosched()

	// Move clock forward to just before the sleep duration.
	clock.Add(9 * time.Second)

	if atomic.LoadInt32(&ok) == 1 {
		t.Fatal("too early")
	}

	// Move clock forward to after the sleep duration.
	clock.Add(1 * time.Second)

	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too late")
	}
}

// Ensure that the mock's Tick channel sends at the correct time.
func TestMock_Tick(t *testing.T) {
	var (
		n     int32
		clock = clock.NewMock()
	)

	// Create a channel to increment every 10 seconds.
	go func() {
		tick := clock.Tick(10 * time.Second)

		for {
			<-tick
			atomic.AddInt32(&n, 1)
		}
	}()

	internal.Gosched()

	// Move clock forward to just before the first tick.
	clock.Add(9 * time.Second)

	if atomic.LoadInt32(&n) != 0 {
		t.Fatalf("expected 0, got %d", n)
	}

	// Move clock forward to the start of the first tick.
	clock.Add(1 * time.Second)

	if atomic.LoadInt32(&n) != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	// Move clock forward over several ticks.
	clock.Add(30 * time.Second)

	if atomic.LoadInt32(&n) != 4 {
		t.Fatalf("expected 4, got %d", n)
	}
}

// Ensure that the mock's Ticker channel sends at the correct time.
func TestMock_Ticker(t *testing.T) {
	var (
		n     int32
		clock = clock.NewMock()
	)

	// Create a channel to increment every microsecond.
	go func() {
		ticker := clock.Ticker(1 * time.Microsecond)

		for {
			<-ticker.C
			atomic.AddInt32(&n, 1)
		}
	}()

	internal.Gosched()

	// Move clock forward.
	clock.Add(10 * time.Microsecond)

	if atomic.LoadInt32(&n) != 10 {
		t.Fatalf("unexpected: %d", n)
	}
}

// Ensure that the mock's Ticker channel won't block if not read from.
func TestMock_Ticker_Overflow(t *testing.T) {
	clock := clock.NewMock()
	ticker := clock.Ticker(1 * time.Microsecond)

	clock.Add(10 * time.Microsecond)
	ticker.Stop()
}

// Ensure that the mock's Ticker can be stopped.
func TestMock_Ticker_Stop(t *testing.T) {
	var (
		n     int32
		clock = clock.NewMock()
	)

	// Create a channel to increment every second.
	ticker := clock.Ticker(1 * time.Second)

	go func() {
		for {
			<-ticker.C
			atomic.AddInt32(&n, 1)
		}
	}()

	internal.Gosched()

	// Move clock forward.
	clock.Add(5 * time.Second)

	if atomic.LoadInt32(&n) != 5 {
		t.Fatalf("expected 5, got: %d", n)
	}

	ticker.Stop()

	// Move clock forward again.
	clock.Add(5 * time.Second)

	if atomic.LoadInt32(&n) != 5 {
		t.Fatalf("still expected 5, got: %d", n)
	}
}

func TestMock_Ticker_Reset(t *testing.T) {
	var (
		n     int32
		clock = clock.NewMock()
	)

	ticker := clock.Ticker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C
			atomic.AddInt32(&n, 1)
		}
	}()
	internal.Gosched()

	// Move clock forward.
	clock.Add(10 * time.Second)

	if atomic.LoadInt32(&n) != 2 {
		t.Fatalf("expected 2, got: %d", n)
	}

	clock.Add(4 * time.Second)
	ticker.Reset(5 * time.Second)

	// Advance the remaining second
	clock.Add(1 * time.Second)

	if atomic.LoadInt32(&n) != 2 {
		t.Fatalf("expected 2, got: %d", n)
	}

	// Advance the remaining 4 seconds from the previous tick
	clock.Add(4 * time.Second)

	if atomic.LoadInt32(&n) != 3 {
		t.Fatalf("expected 3, got: %d", n)
	}
}

func TestMock_Ticker_Stop_Reset(t *testing.T) {
	var (
		n     int32
		clock = clock.NewMock()
	)

	ticker := clock.Ticker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C
			atomic.AddInt32(&n, 1)
		}
	}()
	internal.Gosched()

	// Move clock forward.
	clock.Add(10 * time.Second)

	if atomic.LoadInt32(&n) != 2 {
		t.Fatalf("expected 2, got: %d", n)
	}

	ticker.Stop()

	// Move clock forward again.
	clock.Add(5 * time.Second)

	if atomic.LoadInt32(&n) != 2 {
		t.Fatalf("still expected 2, got: %d", n)
	}

	ticker.Reset(2 * time.Second)

	// Advance the remaining 2 seconds
	clock.Add(2 * time.Second)

	if atomic.LoadInt32(&n) != 3 {
		t.Fatalf("expected 3, got: %d", n)
	}

	// Advance another 2 seconds
	clock.Add(2 * time.Second)

	if atomic.LoadInt32(&n) != 4 {
		t.Fatalf("expected 4, got: %d", n)
	}
}

// Ensure that multiple tickers can be used together.
func TestMock_Ticker_Multi(t *testing.T) {
	var (
		n     int32
		clock = clock.NewMock()
	)

	go func() {
		a := clock.Ticker(1 * time.Microsecond)
		b := clock.Ticker(3 * time.Microsecond)

		for {
			select {
			case <-a.C:
				atomic.AddInt32(&n, 1)
			case <-b.C:
				atomic.AddInt32(&n, 100)
			}
		}
	}()

	internal.Gosched()

	// Move clock forward.
	clock.Add(10 * time.Microsecond)
	internal.Gosched()

	if atomic.LoadInt32(&n) != 310 {
		t.Fatalf("unexpected: %d", n)
	}
}

func ExampleMock_After() {
	var ( // Create a new mock clock.
		clock = clock.NewMock()
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
		clock = clock.NewMock()
		count counter
	)

	count.incr()

	// Execute a function after 10 mock seconds.
	clock.AfterFunc(10*time.Second, func() {
		count.incr()
	})
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
		clock = clock.NewMock()
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
		clock = clock.NewMock()
		count counter
	)

	ready := make(chan struct{})
	// Increment count every mock second.
	go func() {
		ticker := clock.Ticker(1 * time.Second)

		close(ready)

		for {
			<-ticker.C
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
		clock = clock.NewMock()
		count counter
	)

	ready := make(chan struct{})
	// Increment count after a mock second.
	go func() {
		timer := clock.Timer(1 * time.Second)

		close(ready)

		<-timer.C
		count.incr()
	}()
	<-ready

	// Move the clock forward 10 seconds and print the new value.
	clock.Add(10 * time.Second)
	fmt.Printf("Count is %d after 10 seconds\n", count.get())

	// Output:
	// Count is 1 after 10 seconds
}

func TestMock_ReentrantDeadlock(t *testing.T) {
	mockedClock := clock.NewMock()
	timer20 := mockedClock.Timer(20 * time.Second)

	go func() {
		v := <-timer20.C
		panic(fmt.Sprintf("timer should not have ticked: %v", v))
	}()

	mockedClock.AfterFunc(10*time.Second, func() {
		timer20.Stop()
	})

	mockedClock.Add(15 * time.Second)
	mockedClock.Add(15 * time.Second)
}

func TestMock_AddAfterFuncRace(t *testing.T) {
	// start blocks the goroutines in this test
	// until we're ready for them to race.
	start := make(chan struct{})

	var wg sync.WaitGroup

	mockedClock := clock.NewMock()

	var calls counter
	defer func() {
		if calls.get() == 0 {
			t.Errorf("AfterFunc did not call the function")
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		<-start

		mockedClock.AfterFunc(time.Millisecond, func() {
			calls.incr()
		})
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		<-start

		mockedClock.Add(time.Millisecond)
		mockedClock.Add(time.Millisecond)
	}()

	close(start) // unblock the goroutines
	wg.Wait()    // and wait for them
}

func TestMock_AfterRace(t *testing.T) {
	const num = 10

	var (
		mock     = clock.NewMock()
		finished atomic.Int32
	)

	for i := 0; i < num; i++ {
		go func() {
			<-mock.After(1 * time.Millisecond)
			finished.Add(1)
		}()
	}

	for finished.Load() < num {
		mock.Add(time.Second)
		internal.Gosched()
	}
}
