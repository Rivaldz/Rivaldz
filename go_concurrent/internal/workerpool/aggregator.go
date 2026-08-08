package workerpool

type AggregatedData struct {
	CategoryCount     map[string]int
	CategoryDateCount map[string]map[string]int
	TopCategory       string
	TopCategoryCount  int
	TopDate           string
	TopDateCount      int
}

func Aggregate(resultCh <-chan *Result) *AggregatedData {
	finalCatCount := make(map[string]int)
	finalCatDateCount := make(map[string]map[string]int)
	
	for res := range resultCh {
		for cat, count := range res.CategoryCount {
			finalCatCount[cat] += count
		}
		
		for cat, dates := range res.CategoryDateCount {
			if finalCatDateCount[cat] == nil {
				finalCatDateCount[cat] = make(map[string]int)
			}
			for date, count := range dates {
				finalCatDateCount[cat][date] += count
			}
		}
	}
	
	var topCat string
	var maxCat int
	for cat, count := range finalCatCount {
		if count > maxCat {
			maxCat = count
			topCat = cat
		}
	}
	
	var topDate string
	var maxDate int
	if dates, ok := finalCatDateCount[topCat]; ok {
		for date, count := range dates {
			if count > maxDate {
				maxDate = count
				topDate = date
			}
		}
	}
	
	return &AggregatedData{
		CategoryCount:     finalCatCount,
		CategoryDateCount: finalCatDateCount,
		TopCategory:       topCat,
		TopCategoryCount:  maxCat,
		TopDate:           topDate,
		TopDateCount:      maxDate,
	}
}
