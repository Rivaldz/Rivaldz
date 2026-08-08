package workerpool

import (
	"testing"
)

func TestAggregate_MergeMaps(t *testing.T) {
	t.Parallel()
	resultCh := make(chan *Result, 2)
	
	res1 := &Result{
		CategoryCount: map[string]int{"Cat1": 10, "Cat2": 5},
		CategoryDateCount: map[string]map[string]int{
			"Cat1": {"2023-01-01": 10},
			"Cat2": {"2023-01-01": 5},
		},
	}
	res2 := &Result{
		CategoryCount: map[string]int{"Cat1": 15, "Cat3": 20},
		CategoryDateCount: map[string]map[string]int{
			"Cat1": {"2023-01-02": 15},
			"Cat3": {"2023-01-01": 20},
		},
	}

	resultCh <- res1
	resultCh <- res2
	close(resultCh)

	agg := Aggregate(resultCh)

	if agg.CategoryCount["Cat1"] != 25 {
		t.Errorf("expected 25 for Cat1, got %d", agg.CategoryCount["Cat1"])
	}
	if agg.CategoryCount["Cat2"] != 5 {
		t.Errorf("expected 5 for Cat2, got %d", agg.CategoryCount["Cat2"])
	}
	if agg.CategoryCount["Cat3"] != 20 {
		t.Errorf("expected 20 for Cat3, got %d", agg.CategoryCount["Cat3"])
	}
}

func TestAggregate_FindMax(t *testing.T) {
	t.Parallel()
	resultCh := make(chan *Result, 1)
	
	res1 := &Result{
		CategoryCount: map[string]int{"Cat1": 10, "Cat2": 50},
		CategoryDateCount: map[string]map[string]int{
			"Cat1": {"2023-01-01": 10},
			"Cat2": {"2023-01-01": 20, "2023-01-02": 30},
		},
	}

	resultCh <- res1
	close(resultCh)

	agg := Aggregate(resultCh)

	if agg.TopCategory != "Cat2" {
		t.Errorf("expected TopCategory Cat2, got %s", agg.TopCategory)
	}
	if agg.TopCategoryCount != 50 {
		t.Errorf("expected TopCategoryCount 50, got %d", agg.TopCategoryCount)
	}
	if agg.TopDate != "2023-01-02" {
		t.Errorf("expected TopDate 2023-01-02, got %s", agg.TopDate)
	}
	if agg.TopDateCount != 30 {
		t.Errorf("expected TopDateCount 30, got %d", agg.TopDateCount)
	}
}

func TestAggregate_SingleCategory(t *testing.T) {
	t.Parallel()
	resultCh := make(chan *Result, 1)
	
	res1 := &Result{
		CategoryCount: map[string]int{"Cat1": 10},
		CategoryDateCount: map[string]map[string]int{
			"Cat1": {"2023-01-01": 10},
		},
	}

	resultCh <- res1
	close(resultCh)

	agg := Aggregate(resultCh)

	if agg.TopCategory != "Cat1" {
		t.Errorf("expected TopCategory Cat1, got %s", agg.TopCategory)
	}
	if agg.TopCategoryCount != 10 {
		t.Errorf("expected TopCategoryCount 10, got %d", agg.TopCategoryCount)
	}
}

func TestAggregate_EmptyInput(t *testing.T) {
	t.Parallel()
	resultCh := make(chan *Result)
	close(resultCh)

	agg := Aggregate(resultCh)

	if agg.TopCategory != "" {
		t.Errorf("expected empty TopCategory, got %s", agg.TopCategory)
	}
	if agg.TopCategoryCount != 0 {
		t.Errorf("expected TopCategoryCount 0, got %d", agg.TopCategoryCount)
	}
	if len(agg.CategoryCount) != 0 {
		t.Errorf("expected empty CategoryCount map")
	}
}
