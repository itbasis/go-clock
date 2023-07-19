package internal

import "time"

// Sleep momentarily so that other goroutines can process.
func Gosched() { time.Sleep(1 * time.Millisecond) }
