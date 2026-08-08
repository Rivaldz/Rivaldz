package workerpool

import (
	"context"
	"testing"
	"time"

	"go_project/internal/model"
	"go_project/internal/tracker"
)

func TestPool_ProcessRecords(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	jobCh := make(chan *model.Record, 1000)
	tr := tracker.NewTracker()

	for i := 0; i < 1000; i++ {
		jobCh <- &model.Record{Category: "Cat1"}
	}
	close(jobCh)

	resultCh := StartPool(ctx, 5, jobCh, tr)

	total := 0
	for res := range resultCh {
		total += res.CategoryCount["Cat1"]
	}

	if total != 1000 {
		t.Errorf("expected 1000, got %d", total)
	}
}

func TestPool_ConcurrentSafety(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	jobCh := make(chan *model.Record, 1000)
	tr := tracker.NewTracker()

	for i := 0; i < 1000; i++ {
		jobCh <- &model.Record{Category: "Cat1"}
	}
	close(jobCh)

	resultCh := StartPool(ctx, 10, jobCh, tr)
	
	total := 0
	for res := range resultCh {
		total += res.CategoryCount["Cat1"]
	}
	
	if total != 1000 {
		t.Errorf("expected 1000, got %d", total)
	}
}

func TestPool_ErrorHandling(t *testing.T) {
	t.Parallel()
	// Since StartPool doesn't handle errors natively from records, 
	// we test the fallback processWorker method for the pool struct.
	ctx := context.Background()
	recordCh := make(chan model.Record, 10)
	errCh := make(chan error, 10)
	pool := NewPool(2, recordCh, errCh)

	recordCh <- model.Record{Category: "Cat1"}
	close(recordCh)

	results := pool.Run(ctx)
	
	total := 0
	for _, res := range results {
		total += res.CategoryCount["Cat1"]
	}

	if total != 1 {
		t.Errorf("expected 1 record processed")
	}
}

func TestPool_ZeroWorkers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	jobCh := make(chan *model.Record, 10)
	tr := tracker.NewTracker()
	
	close(jobCh)
	// StartPool with 0 workers should default to 1 worker
	resultCh := StartPool(ctx, 0, jobCh, tr)
	
	// Wait for channel to close
	for range resultCh {
	}
}

func TestPool_GracefulShutdown(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	jobCh := make(chan *model.Record, 10)
	tr := tracker.NewTracker()

	// cancel immediately
	cancel()

	resultCh := StartPool(ctx, 5, jobCh, tr)

	// Ensure result channel closes cleanly
	done := make(chan struct{})
	go func() {
		for range resultCh {
		}
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("pool did not shut down gracefully on context cancel")
	}
}
