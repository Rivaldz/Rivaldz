package csvreader

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestGenerateCSV_RowCount(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	gen := NewGenerator(tmpDir, 1, 100)
	
	files, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f, err := os.Open(files[0])
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	// 100 rows + 1 header
	if len(records) != 101 {
		t.Errorf("expected 101 rows, got %d", len(records))
	}
}

func TestGenerateCSV_Columns(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	gen := NewGenerator(tmpDir, 1, 10)
	files, _ := gen.Generate()

	f, _ := os.Open(files[0])
	defer f.Close()
	reader := csv.NewReader(f)
	
	header, _ := reader.Read()
	if len(header) != 7 {
		t.Errorf("expected 7 header columns, got %d", len(header))
	}

	expectedHeader := []string{"id", "name", "email", "amount", "category", "date", "status"}
	for i, h := range header {
		if h != expectedHeader[i] {
			t.Errorf("expected header %s, got %s", expectedHeader[i], h)
		}
	}
}

func TestGenerateCSV_DataTypes(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	gen := NewGenerator(tmpDir, 1, 5)
	files, _ := gen.Generate()

	f, _ := os.Open(files[0])
	defer f.Close()
	reader := csv.NewReader(f)
	
	_, _ = reader.Read() // skip header
	row, _ := reader.Read()
	
	// ID numeric
	if _, err := strconv.Atoi(row[0]); err != nil {
		t.Errorf("expected numeric ID, got %s: %v", row[0], err)
	}
	
	// amount float
	if _, err := strconv.ParseFloat(row[3], 64); err != nil {
		t.Errorf("expected float amount, got %s: %v", row[3], err)
	}
	
	// date format YYYY-MM-DD
	if _, err := time.Parse("2006-01-02", row[5]); err != nil {
		t.Errorf("expected date YYYY-MM-DD, got %s: %v", row[5], err)
	}
}

func TestGenerateCSV_MultipleFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	gen := NewGenerator(tmpDir, 3, 100) // 100 / 3 = 33, 33, 34
	
	files, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	totalRows := 0
	for _, file := range files {
		f, _ := os.Open(file)
		reader := csv.NewReader(f)
		records, _ := reader.ReadAll()
		f.Close()
		
		totalRows += len(records) - 1 // skip header
		
		expectedHeader := []string{"id", "name", "email", "amount", "category", "date", "status"}
		for i, h := range records[0] {
			if h != expectedHeader[i] {
				t.Errorf("expected header %s, got %s", expectedHeader[i], h)
			}
		}
	}

	if totalRows != 100 {
		t.Errorf("expected 100 total rows, got %d", totalRows)
	}
}
