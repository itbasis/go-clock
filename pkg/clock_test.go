package pkg_test

import (
	"sync"
	"testing"
	"time"

	"github.com/itbasis/go-clock/v2/internal/impl"
	"github.com/itbasis/go-clock/v2/internal/mock"
)

// Ensure that the clock's After channel sends at the correct time.
func TestClock_After(t *testing.T) {
	start := time.Now()
	<-impl.NewClock().After(20 * time.Millisecond)
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

	impl.NewClock().AfterFunc(
		20*time.Millisecond, func() {
			ok = true
			wg.Done()
		},
	)
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
	b := impl.NewClock().Now().Round(time.Second)

	if !a.Equal(b) {
		t.Errorf("not equal: %s != %s", a, b)
	}
}

// Ensure that the clock sleeps for the appropriate amount of time.
func TestClock_Sleep(t *testing.T) {
	start := time.Now()
	impl.NewClock().Sleep(20 * time.Millisecond)
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock ticks correctly.
func TestClock_Tick(t *testing.T) {
	start := time.Now()
	c := impl.NewClock().Tick(20 * time.Millisecond)
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
	ticker := impl.NewClock().Ticker(50 * time.Millisecond)
	<-ticker.Chan()
	<-ticker.Chan()

	dur := time.Since(start)

	if dur < 100*time.Millisecond || dur > 200*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock's ticker can stop correctly.
func TestClock_Ticker_Stp(t *testing.T) {
	ticker := impl.NewClock().Ticker(20 * time.Millisecond)
	<-ticker.Chan()
	ticker.Stop()
	select {
	case <-ticker.Chan():
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's ticker can reset correctly.
func TestClock_Ticker_Rst(t *testing.T) {
	start := time.Now()
	ticker := impl.NewClock().Ticker(20 * time.Millisecond)
	<-ticker.Chan()
	ticker.Reset(5 * time.Millisecond)
	<-ticker.Chan()

	dur := time.Since(start)

	if dur >= 30*time.Millisecond {
		t.Fatal("took more than 30ms")
	}

	ticker.Stop()
}

// Ensure that the clock's ticker can stop and then be reset correctly.
func TestClock_Ticker_Stop_Rst(t *testing.T) {
	start := time.Now()
	ticker := impl.NewClock().Ticker(20 * time.Millisecond)
	<-ticker.Chan()
	ticker.Stop()

	select {
	case <-ticker.Chan():
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
	ticker.Reset(5 * time.Millisecond)
	<-ticker.Chan()

	dur := time.Since(start)

	if dur >= 60*time.Millisecond {
		t.Fatal("took more than 60ms")
	}

	ticker.Stop()
}

// Ensure that the clock's timer waits correctly.
func TestClock_Timer(t *testing.T) {
	start := time.Now()
	timer := impl.NewClock().Timer(20 * time.Millisecond)
	<-timer.Chan()

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
	timer := impl.NewClock().Timer(20 * time.Millisecond)

	if !timer.Stop() {
		t.Fatal("timer not running")
	}

	if timer.Stop() {
		t.Fatal("timer wasn't cancelled")
	}

	select {
	case <-timer.Chan():
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's timer can be reset.
func TestClock_Timer_Reset(t *testing.T) {
	start := time.Now()
	timer := impl.NewClock().Timer(10 * time.Millisecond)

	if !timer.Reset(20 * time.Millisecond) {
		t.Fatal("timer not running")
	}

	<-timer.Chan()

	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

func TestClock_NegativeDuration(t *testing.T) {
	clock := mock.NewMock()
	timer := clock.Timer(-time.Second)
	select {
	case <-timer.Chan():
	default:
		t.Fatal("timer should have fired immediately")
	}
}

// Ensure reset can be called immediately after reading channel
func TestClock_Timer_Reset_Unlock(t *testing.T) {
	var (
		clock = mock.NewMock()
		timer = clock.Timer(1 * time.Second)
		wg    sync.WaitGroup
	)

	wg.Add(1)

	go func() {
		defer wg.Done()

		select { //nolint:gosimple
		case <-timer.Chan():
			println("case reset")
			timer.Reset(1 * time.Second)
		}

		select { //nolint:gosimple
		case <-timer.Chan():
			println("case read")
		}
	}()

	clock.Add(2 * time.Second)
	wg.Wait()
}
