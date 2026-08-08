package workerpool

import (
	"context"
	"sync"

	"go_project/internal/model"
	"go_project/internal/tracker"
)

type Result struct {
	CategoryCount     map[string]int
	CategoryDateCount map[string]map[string]int
}

func StartPool(ctx context.Context, workers int, jobCh <-chan *model.Record, tr *tracker.Tracker) <-chan *Result {
	if workers <= 0 {
		workers = 1
	}
	resultCh := make(chan *Result, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker(ctx, jobCh, resultCh, tr)
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}

type Pool struct {
	Workers int
	RecordCh <-chan model.Record
	ErrCh    chan<- error
}

func NewPool(workers int, recordCh <-chan model.Record, errCh chan<- error) *Pool {
	if workers <= 0 {
		workers = 1
	}
	return &Pool{
		Workers: workers,
		RecordCh: recordCh,
		ErrCh:   errCh,
	}
}

func (p *Pool) Run(ctx context.Context) []Result {
	results := make([]Result, p.Workers)
	var wg sync.WaitGroup

	for i := 0; i < p.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			results[workerID] = processWorker(ctx, workerID, p.RecordCh, p.ErrCh)
		}(i)
	}

	wg.Wait()

	actual := 0
	for _, r := range results {
		if r.CategoryCount != nil || r.CategoryDateCount != nil {
			actual++
		}
	}
	if actual == 0 {
		return nil
	}

	return results[:actual]
}

func processWorker(ctx context.Context, workerID int, recordCh <-chan model.Record, errCh chan<- error) Result {
	result := Result{
		CategoryCount:     make(map[string]int),
		CategoryDateCount: make(map[string]map[string]int),
	}

	defer func() {
		if r := recover(); r != nil {
			select {
			case errCh <- &WorkerPanicError{WorkerID: workerID, Panic: r}:
			case <-ctx.Done():
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return result
		case record, ok := <-recordCh:
			if !ok {
				return result
			}
			result.CategoryCount[record.Category]++

			if result.CategoryDateCount[record.Category] == nil {
				result.CategoryDateCount[record.Category] = make(map[string]int)
			}
			result.CategoryDateCount[record.Category][record.Date]++
		}
	}
}

type WorkerPanicError struct {
	WorkerID int
	Panic    interface{}
}

func (e *WorkerPanicError) Error() string {
	return "worker panicked"
}
