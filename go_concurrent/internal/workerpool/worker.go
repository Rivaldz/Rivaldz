package workerpool

import (
	"context"
	"log"
	"go_project/internal/model"
	"go_project/internal/tracker"
)

func worker(ctx context.Context, jobCh <-chan *model.Record, resultCh chan<- *Result, tr *tracker.Tracker) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker recovered from panic: %v", r)
		}
	}()

	catCount := make(map[string]int)
	catDateCount := make(map[string]map[string]int)

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobCh:
			if !ok {
				resultCh <- &Result{
					CategoryCount:     catCount,
					CategoryDateCount: catDateCount,
				}
				return
			}
			
			catCount[job.Category]++
			if catDateCount[job.Category] == nil {
				catDateCount[job.Category] = make(map[string]int)
			}
			catDateCount[job.Category][job.Date]++
			
			tr.IncProcessed()
		}
	}
}
