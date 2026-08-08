package csvreader

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func createTempCSV(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadCSV_SingleFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := createTempCSV(t, dir, "data.csv", "id,name,email,amount,category,date,status\n1,A,a@b.c,10.0,Cat1,2023-01-01,ok\n2,B,b@c.d,20.0,Cat2,2023-01-02,ok\n")

	reader := NewReader(10)
	ctx := context.Background()
	reader.ReadFiles(ctx, []string{file})

	count := 0
	for range reader.RecordCh {
		count++
	}

	if count != 2 {
		t.Errorf("expected 2 records, got %d", count)
	}
}

func TestReadCSV_MultipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file1 := createTempCSV(t, dir, "data1.csv", "id,name,email,amount,category,date,status\n1,A,a@b.c,10.0,Cat1,2023-01-01,ok\n")
	file2 := createTempCSV(t, dir, "data2.csv", "id,name,email,amount,category,date,status\n2,B,b@c.d,20.0,Cat2,2023-01-02,ok\n")
	file3 := createTempCSV(t, dir, "data3.csv", "id,name,email,amount,category,date,status\n3,C,c@d.e,30.0,Cat3,2023-01-03,ok\n")

	reader := NewReader(10)
	ctx := context.Background()
	reader.ReadFiles(ctx, []string{file1, file2, file3})

	count := 0
	for range reader.RecordCh {
		count++
	}

	if count != 3 {
		t.Errorf("expected 3 records, got %d", count)
	}
}

func TestReadCSV_EmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := createTempCSV(t, dir, "empty.csv", "")

	reader := NewReader(10)
	ctx := context.Background()
	reader.ReadFiles(ctx, []string{file})

	count := 0
	for range reader.RecordCh {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 records, got %d", count)
	}
	
	errCount := 0
	for err := range reader.ErrCh {
		if err != nil {
			errCount++
		}
	}
	if errCount != 1 {
		t.Errorf("expected 1 error for EOF/read header on empty file, got %d", errCount)
	}
}

func TestReadCSV_MalformedRows(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := "id,name,email,amount,category,date,status\n" +
		"1,A,a@b.c,10.0,Cat1,2023-01-01,ok\n" +
		"bad_id,B,b@c.d,20.0,Cat2,2023-01-02,ok\n" +
		"3,C,c@d.e,bad_amount,Cat3,2023-01-03,ok\n" +
		"4,D\n" // missing columns

	file := createTempCSV(t, dir, "data.csv", content)

	reader := NewReader(10)
	ctx := context.Background()
	reader.ReadFiles(ctx, []string{file})

	recordCount := 0
	for range reader.RecordCh {
		recordCount++
	}

	if recordCount != 1 {
		t.Errorf("expected 1 valid record, got %d", recordCount)
	}

	errCount := 0
	for err := range reader.ErrCh {
		if err != nil {
			errCount++
		}
	}

	// 3 errors: bad_id, bad_amount, missing columns
	if errCount != 3 {
		t.Errorf("expected 3 errors, got %d", errCount)
	}
}

func TestReadCSV_FileNotFound(t *testing.T) {
	t.Parallel()
	reader := NewReader(10)
	ctx := context.Background()
	reader.ReadFiles(ctx, []string{"nonexistent.csv"})

	errCount := 0
	for err := range reader.ErrCh {
		if err != nil {
			errCount++
		}
	}

	if errCount != 1 {
		t.Errorf("expected 1 error for file not found, got %d", errCount)
	}
}

func TestReadCSV_ContextCancellation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := createTempCSV(t, dir, "data.csv", "id,name,email,amount,category,date,status\n1,A,a@b.c,10.0,Cat1,2023-01-01,ok\n")

	reader := NewReader(10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	reader.ReadFiles(ctx, []string{file})

	count := 0
	for range reader.RecordCh {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 records due to cancellation, got %d", count)
	}
}

func TestReadCSV_LargeFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	gen := NewGenerator(dir, 1, 10000)
	files, err := gen.Generate()
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	reader := NewReader(100)
	ctx := context.Background()
	reader.ReadFiles(ctx, files)

	count := 0
	for range reader.RecordCh {
		count++
	}

	if count != 10000 {
		t.Errorf("expected 10000 records, got %d", count)
	}
}

func BenchmarkReadCSV(b *testing.B) {
	dir := b.TempDir()
	gen := NewGenerator(dir, 1, 10000)
	files, err := gen.Generate()
	if err != nil {
		b.Fatalf("failed to generate: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := NewReader(1000)
		ctx := context.Background()
		reader.ReadFiles(ctx, files)

		for range reader.RecordCh {
		}
		for range reader.ErrCh {
		}
	}
}
