package tracker

import (
	"sync"
	"testing"
	"time"
)

func TestProgress_IncrementTotal(t *testing.T) {
	t.Parallel()
	tracker := NewTracker()
	for i := 0; i < 5; i++ {
		tracker.IncRead()
		tracker.IncProcessed()
	}

	read, processed, _ := tracker.GetStats()
	if read != 5 {
		t.Errorf("expected read 5, got %d", read)
	}
	if processed != 5 {
		t.Errorf("expected processed 5, got %d", processed)
	}
}

func TestProgress_IncrementErrors(t *testing.T) {
	t.Parallel()
	tracker := NewTracker()
	for i := 0; i < 3; i++ {
		tracker.IncErrors()
	}

	_, _, errors := tracker.GetStats()
	if errors != 3 {
		t.Errorf("expected errors 3, got %d", errors)
	}
}

func TestProgress_ConcurrentInc(t *testing.T) {
	t.Parallel()
	tracker := NewTracker()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.IncRead()
			tracker.IncProcessed()
			tracker.IncErrors()
		}()
	}

	wg.Wait()

	read, processed, errors := tracker.GetStats()
	if read != 100 || processed != 100 || errors != 100 {
		t.Errorf("expected 100 for all, got read=%d, processed=%d, errors=%d", read, processed, errors)
	}
}

func TestProgress_ReportInterval(t *testing.T) {
	t.Parallel()
	tracker := NewTracker()
	tracker.StartPeriodicReport("10ms")
	
	// Wait a bit to let it tick at least once
	time.Sleep(25 * time.Millisecond)
	
	tracker.Stop()
	
	// Ensure it stopped properly
	_, ok := <-tracker.done
	if ok {
		t.Error("expected done channel to be closed")
	}
}

func TestProgress_Summary(t *testing.T) {
	t.Parallel()
	tracker := NewTracker()
	tracker.IncRead()
	tracker.IncProcessed()
	tracker.IncErrors()

	tracker.Stop()

	read, processed, errors := tracker.GetStats()
	if read != 1 || processed != 1 || errors != 1 {
		t.Errorf("expected 1 for all, got read=%d, processed=%d, errors=%d", read, processed, errors)
	}
}
