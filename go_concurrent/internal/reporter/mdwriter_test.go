package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go_project/internal/workerpool"
)

func TestMDWriter_BasicOutput(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.md")

	agg := &workerpool.AggregatedData{
		TopCategory:      "Cat1",
		TopCategoryCount: 100,
		TopDate:          "2023-01-01",
		TopDateCount:     50,
	}

	stats := struct {
		FilesProcessed int
		TotalRead      int64
		TotalProcessed int64
		TotalErrors    int64
		Duration       time.Duration
		Workers        int
	}{
		FilesProcessed: 3,
		TotalRead:      1000,
		TotalProcessed: 900,
		TotalErrors:    100,
		Duration:       2 * time.Second,
		Workers:        5,
	}

	err := WriteReport(outputPath, agg, stats)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("expected file to exist")
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "# CSV Processing Report") {
		t.Errorf("expected title in markdown")
	}
	if !strings.Contains(strContent, "Total files processed: 3") {
		t.Errorf("expected files processed to be 3")
	}
	if !strings.Contains(strContent, "Cat1` (100 records)") {
		t.Errorf("expected top category Cat1 with 100 records")
	}
	if !strings.Contains(strContent, "2023-01-01` (50 records)") {
		t.Errorf("expected peak date 2023-01-01 with 50 records")
	}
}

func TestMDWriter_FilePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.md")

	agg := &workerpool.AggregatedData{}
	stats := struct {
		FilesProcessed int
		TotalRead      int64
		TotalProcessed int64
		TotalErrors    int64
		Duration       time.Duration
		Workers        int
	}{Duration: 1 * time.Second}

	_ = WriteReport(outputPath, agg, stats)

	f, err := os.Open(outputPath)
	if err != nil {
		t.Errorf("expected file to be readable, got %v", err)
	}
	f.Close()
}

func TestMDWriter_Overwrite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.md")

	agg := &workerpool.AggregatedData{}
	stats := struct {
		FilesProcessed int
		TotalRead      int64
		TotalProcessed int64
		TotalErrors    int64
		Duration       time.Duration
		Workers        int
	}{Duration: 1 * time.Second}

	// First write
	err1 := WriteReport(outputPath, agg, stats)
	if err1 != nil {
		t.Fatal(err1)
	}
	info1, _ := os.Stat(outputPath)
	modTime1 := info1.ModTime()

	// Wait a bit to ensure ModTime is different (or just check the content changed)
	time.Sleep(10 * time.Millisecond)

	// Second write with different content
	stats.FilesProcessed = 999
	err2 := WriteReport(outputPath, agg, stats)
	if err2 != nil {
		t.Fatal(err2)
	}

	info2, _ := os.Stat(outputPath)
	if !info2.ModTime().After(modTime1) && info2.ModTime() != modTime1 {
		// Not exactly strict if filesystem doesn't support fine grained time,
		// but checking content is better
	}

	content, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(content), "Total files processed: 999") {
		t.Errorf("expected file to be overwritten with new content")
	}
}

func TestMDWriter_EmptyData(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report_empty.md")

	agg := &workerpool.AggregatedData{}
	stats := struct {
		FilesProcessed int
		TotalRead      int64
		TotalProcessed int64
		TotalErrors    int64
		Duration       time.Duration
		Workers        int
	}{Duration: 1 * time.Second}

	err := WriteReport(outputPath, agg, stats)
	if err != nil {
		t.Fatalf("expected no error for empty data, got %v", err)
	}

	content, _ := os.ReadFile(outputPath)
	strContent := string(content)
	
	if !strings.Contains(strContent, "Total files processed: 0") {
		t.Errorf("expected 0 files processed")
	}
}
