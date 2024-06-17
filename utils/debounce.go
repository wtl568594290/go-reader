package utils

import (
	"sync"
	"time"
)

type Debouncer struct {
	duration time.Duration
	timer    *time.Timer
	mu       sync.Mutex
}

func NewDebouncer(ms int) *Debouncer {
	return &Debouncer{
		duration: time.Duration(ms) * time.Millisecond,
	}
}

func (d *Debouncer) Debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.duration, f)
}
