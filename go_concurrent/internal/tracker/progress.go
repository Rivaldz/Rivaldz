package tracker

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Tracker struct {
	totalRead      atomic.Int64
	totalProcessed atomic.Int64
	totalErrors    atomic.Int64
	done           chan struct{}
}

func NewTracker() *Tracker {
	return &Tracker{
		done: make(chan struct{}),
	}
}

func (t *Tracker) IncRead() {
	t.totalRead.Add(1)
}

func (t *Tracker) IncProcessed() {
	t.totalProcessed.Add(1)
}

func (t *Tracker) IncErrors() {
	t.totalErrors.Add(1)
}

func (t *Tracker) GetStats() (read, processed, errors int64) {
	return t.totalRead.Load(), t.totalProcessed.Load(), t.totalErrors.Load()
}

func (t *Tracker) StartPeriodicReport(intervalStr string) {
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = time.Second
	}
	
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				r, p, e := t.GetStats()
				fmt.Printf("[Progress] Read: %d | Processed: %d | Errors: %d\n", r, p, e)
			case <-t.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (t *Tracker) Stop() {
	close(t.done)
}
