package model

import (
	"fmt"
	"strconv"
	"strings"
)

type Record struct {
	ID       int
	Name     string
	Email    string
	Amount   float64
	Category string
	Date     string
	Status   string
}

func ParseRecord(fields []string) (Record, error) {
	if len(fields) < 7 {
		return Record{}, fmt.Errorf("expected 7 fields, got %d", len(fields))
	}

	id, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil {
		return Record{}, fmt.Errorf("invalid id: %w", err)
	}

	amount, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
	if err != nil {
		return Record{}, fmt.Errorf("invalid amount: %w", err)
	}

	return Record{
		ID:       id,
		Name:     strings.TrimSpace(fields[1]),
		Email:    strings.TrimSpace(fields[2]),
		Amount:   amount,
		Category: strings.TrimSpace(fields[4]),
		Date:     strings.TrimSpace(fields[5]),
		Status:   strings.TrimSpace(fields[6]),
	}, nil
}
