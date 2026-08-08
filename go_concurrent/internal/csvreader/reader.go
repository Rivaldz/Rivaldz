package csvreader

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	"go_project/internal/model"
)

type Reader struct {
	ErrCh    chan error
	RecordCh chan model.Record
}

func NewReader(jobBuf int) *Reader {
	return &Reader{
		ErrCh:    make(chan error, 100),
		RecordCh: make(chan model.Record, jobBuf),
	}
}

func (r *Reader) ReadFiles(ctx context.Context, files []string) error {
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			r.readFile(ctx, f)
		}(file)
	}

	go func() {
		wg.Wait()
		close(r.RecordCh)
		close(r.ErrCh)
	}()

	return nil
}

func (r *Reader) readFile(ctx context.Context, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		select {
		case r.ErrCh <- fmt.Errorf("open %s: %w", filename, err):
		case <-ctx.Done():
		}
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true

	if _, err := reader.Read(); err != nil {
		select {
		case r.ErrCh <- fmt.Errorf("read header %s: %w", filename, err):
		case <-ctx.Done():
		}
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fields, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			select {
			case r.ErrCh <- fmt.Errorf("read row %s: %w", filename, err):
			case <-ctx.Done():
				return
			}
			continue
		}

		record, err := model.ParseRecord(fields)
		if err != nil {
			select {
			case r.ErrCh <- fmt.Errorf("parse %s: %w", filename, err):
			case <-ctx.Done():
				return
			}
			continue
		}

		select {
		case r.RecordCh <- record:
		case <-ctx.Done():
			return
		}
	}
}
