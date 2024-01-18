package mock

// ClockTimers represents a list of sortable timers.
type clockTickers []clockTicker

func (a clockTickers) Len() int           { return len(a) }
func (a clockTickers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a clockTickers) Less(i, j int) bool { return a[i].Next().Before(a[j].Next()) }
