package csvreader

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	categories = []string{"Electronics", "Fashion", "Food", "Sports", "Books", "Home", "Garden", "Toys", "Health", "Automotive"}
	statuses   = []string{"active", "inactive", "pending", "cancelled"}
	firstNames = []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry", "Iris", "Jack",
		"Kate", "Leo", "Mia", "Noah", "Olivia", "Paul", "Quinn", "Rose", "Sam", "Tina",
		"Uma", "Victor", "Wendy", "Xander", "Yara", "Zack", "Ava", "Ben", "Clara", "Dan"}
	lastNames = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Wilson", "Taylor"}
	domains   = []string{"gmail.com", "yahoo.com", "outlook.com", "proton.me", "mail.com"}
)

type Generator struct {
	OutputDir string
	NumFiles  int
	TotalRows int
}

func NewGenerator(outputDir string, numFiles int, totalRows int) *Generator {
	return &Generator{
		OutputDir: outputDir,
		NumFiles:  numFiles,
		TotalRows: totalRows,
	}
}

func (g *Generator) Generate() ([]string, error) {
	if err := os.MkdirAll(g.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	rowsPerFile := g.TotalRows / g.NumFiles
	remainder := g.TotalRows % g.NumFiles

	files := make([]string, g.NumFiles)
	errCh := make(chan error, g.NumFiles)
	var wg sync.WaitGroup

	for i := 0; i < g.NumFiles; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			count := rowsPerFile
			if idx < remainder {
				count++
			}

			filename := filepath.Join(g.OutputDir, fmt.Sprintf("data_%d.csv", idx+1))
			if err := g.generateFile(filename, count, idx*rowsPerFile+1); err != nil {
				errCh <- fmt.Errorf("file %s: %w", filename, err)
				return
			}
			files[idx] = filename
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

func (g *Generator) generateFile(filename string, count, startID int) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"id", "name", "email", "amount", "category", "date", "status"})

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(count)))

	for i := 0; i < count; i++ {
		record := generateRecord(rng, startID+i)
		if err := w.Write(record); err != nil {
			return err
		}

		if i%5000 == 0 {
			w.Flush()
		}
	}

	return nil
}

func generateRecord(rng *rand.Rand, id int) []string {
	firstName := firstNames[rng.Intn(len(firstNames))]
	lastName := lastNames[rng.Intn(len(lastNames))]
	email := fmt.Sprintf("%s.%s%d@%s", firstName, lastName, rng.Intn(1000), domains[rng.Intn(len(domains))])

	year := 2025 + rng.Intn(2)
	month := 1 + rng.Intn(12)
	day := 1 + rng.Intn(28)
	date := fmt.Sprintf("%d-%02d-%02d", year, month, day)

	return []string{
		strconv.Itoa(id),
		firstName + " " + lastName,
		email,
		fmt.Sprintf("%.2f", rng.Float64()*10000),
		categories[rng.Intn(len(categories))],
		date,
		statuses[rng.Intn(len(statuses))],
	}
}
