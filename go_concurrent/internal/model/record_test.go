package model

import (
	"testing"
)

func TestParseRecord_Valid(t *testing.T) {
	t.Parallel()
	fields := []string{"1", "John Doe", "john@example.com", "100.50", "Electronics", "2023-01-01", "Success"}
	record, err := ParseRecord(fields)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if record.ID != 1 {
		t.Errorf("expected ID 1, got %d", record.ID)
	}
	if record.Name != "John Doe" {
		t.Errorf("expected Name 'John Doe', got %s", record.Name)
	}
	if record.Email != "john@example.com" {
		t.Errorf("expected Email 'john@example.com', got %s", record.Email)
	}
	if record.Amount != 100.50 {
		t.Errorf("expected Amount 100.50, got %f", record.Amount)
	}
	if record.Category != "Electronics" {
		t.Errorf("expected Category 'Electronics', got %s", record.Category)
	}
	if record.Date != "2023-01-01" {
		t.Errorf("expected Date '2023-01-01', got %s", record.Date)
	}
	if record.Status != "Success" {
		t.Errorf("expected Status 'Success', got %s", record.Status)
	}
}

func TestParseRecord_InvalidColumns(t *testing.T) {
	t.Parallel()
	fields := []string{"1", "John Doe"}
	_, err := ParseRecord(fields)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseRecord_BadInt(t *testing.T) {
	t.Parallel()
	fields := []string{"abc", "John Doe", "john@example.com", "100.50", "Electronics", "2023-01-01", "Success"}
	_, err := ParseRecord(fields)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseRecord_BadFloat(t *testing.T) {
	t.Parallel()
	fields := []string{"1", "John Doe", "john@example.com", "abc", "Electronics", "2023-01-01", "Success"}
	_, err := ParseRecord(fields)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseRecord_EmptyFields(t *testing.T) {
	t.Parallel()
	fields := []string{"1", "", "", "100.50", "", "", ""}
	record, err := ParseRecord(fields)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if record.Name != "" {
		t.Errorf("expected empty Name, got %s", record.Name)
	}
}
