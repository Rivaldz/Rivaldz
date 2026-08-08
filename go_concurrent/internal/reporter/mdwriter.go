package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"go_project/internal/workerpool"
)

func WriteReport(outputPath string, agg *workerpool.AggregatedData, stats struct {
	FilesProcessed int
	TotalRead      int64
	TotalProcessed int64
	TotalErrors    int64
	Duration       time.Duration
	Workers        int
}) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	
	throughput := float64(stats.TotalProcessed) / stats.Duration.Seconds()
	
	content := fmt.Sprintf(`# CSV Processing Report

**Generated:** %s

## Summary
- Total files processed: %d
- Total rows (read attempt): %d
- Rows processed: %d
- Errors skipped: %d
- Duration: %s

## Analysis
- **Most common category:** `+"`%s`"+` (%d records)
- **Peak date for %s:** `+"`%s`"+` (%d records)

## Processing Details
- Workers: %d
- Throughput: %.0f rows/s
`, 
		time.Now().Format("2006-01-02 15:04:05"),
		stats.FilesProcessed,
		stats.TotalRead,
		stats.TotalProcessed,
		stats.TotalErrors,
		stats.Duration.Round(time.Millisecond),
		agg.TopCategory, agg.TopCategoryCount,
		agg.TopCategory, agg.TopDate, agg.TopDateCount,
		stats.Workers,
		throughput,
	)
	
	_, err = f.WriteString(content)
	return err
}
