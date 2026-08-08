package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	
	"go_project/config"
	"go_project/internal/csvreader"
	"go_project/internal/model"
	"go_project/internal/reporter"
	"go_project/internal/tracker"
	"go_project/internal/workerpool"
)

func main() {
	cfg := config.Load()
	
	fmt.Println("🚀 Starting Concurrent Data Processing...")
	
	if _, err := os.Stat(cfg.CSVDir); os.IsNotExist(err) {
		fmt.Printf("Generating dummy CSV files in %s...\n", cfg.CSVDir)
		gen := csvreader.NewGenerator(cfg.CSVDir, 3, 33334)
		_, err := gen.Generate()
		if err != nil {
			log.Fatalf("Failed to generate dummy data: %v", err)
		}
	} else {
		files, _ := os.ReadDir(cfg.CSVDir)
		hasCSV := false
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".csv" {
				hasCSV = true
				break
			}
		}
		if !hasCSV {
			fmt.Printf("Generating dummy CSV files in %s...\n", cfg.CSVDir)
			gen := csvreader.NewGenerator(cfg.CSVDir, 3, 33334)
			_, err := gen.Generate()
			if err != nil {
				log.Fatalf("Failed to generate dummy data: %v", err)
			}
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	tr := tracker.NewTracker()
	tr.StartPeriodicReport(cfg.ReportInterval.String())
	
	jobCh := make(chan *model.Record, cfg.ChannelBuf)
	
	start := time.Now()
	
	resultCh := workerpool.StartPool(ctx, cfg.Workers, jobCh, tr)
	
	go func() {
		fmt.Println("Reading CSV files...")
		entries, err := os.ReadDir(cfg.CSVDir)
		if err != nil {
			log.Printf("Error reading dir: %v", err)
			close(jobCh)
			return
		}
		var files []string
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".csv" {
				files = append(files, filepath.Join(cfg.CSVDir, e.Name()))
			}
		}

		reader := csvreader.NewReader(cfg.ChannelBuf)
		reader.ReadFiles(ctx, files)
		
		go func() {
			for err := range reader.ErrCh {
				log.Printf("CSV error: %v", err)
				tr.IncErrors()
			}
		}()

		for record := range reader.RecordCh {
			tr.IncRead()
			rec := record
			jobCh <- &rec
		}
		
		close(jobCh)
	}()
	
	aggData := workerpool.Aggregate(resultCh)
	
	duration := time.Since(start)
	tr.Stop()
	
	read, processed, errors := tr.GetStats()
	
	fmt.Printf("\n✅ Processing Completed in %s!\n", duration.Round(time.Millisecond))
	fmt.Printf("Read: %d | Processed: %d | Errors: %d\n", read, processed, errors)
	
	files, _ := os.ReadDir(cfg.CSVDir)
	csvCount := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".csv" {
			csvCount++
		}
	}
	
	stats := struct {
		FilesProcessed int
		TotalRead      int64
		TotalProcessed int64
		TotalErrors    int64
		Duration       time.Duration
		Workers        int
	}{
		FilesProcessed: csvCount,
		TotalRead:      read,
		TotalProcessed: processed,
		TotalErrors:    errors,
		Duration:       duration,
		Workers:        cfg.Workers,
	}
	
	outputPath := filepath.Join(cfg.OutputDir, "report.md")
	err := reporter.WriteReport(outputPath, aggData, stats)
	if err != nil {
		log.Fatalf("Failed to write report: %v", err)
	}
	fmt.Printf("📄 Report saved to %s\n", outputPath)
}
