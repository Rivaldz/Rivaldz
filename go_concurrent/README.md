# Concurrent CSV Data Processing (Go)

A Go application for processing multiple CSV files concurrently using **goroutines**, **channels**, **worker pool pattern**, **graceful error handling**, and **real-time progress tracking**.

## Features

- Read multiple CSV files simultaneously (one goroutine per file)
- Process data with **Worker Pool** (fan-out/fan-in pattern)
- Aggregation: find the most common category and its peak date
- Real-time progress tracking with `atomic.Int64` counters
- Graceful error handling (skip malformed rows, panic recovery, context cancellation)
- Descriptive Markdown output report
- Dummy CSV generator for load testing (100k+ rows)

---

## Architecture & Design Patterns

```
                    ┌─────────────────────────┐
                    │      main.go            │
                    │  ctx + cancel           │
                    │  channels + WaitGroup   │
                    └──────┬──────────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
   [CSV Reader 1]   [CSV Reader 2]   [CSV Reader 3]
   goroutine        goroutine        goroutine
          │                │                │
          └────────────────┼────────────────┘
                           │ jobCh (*Record) buffered channel
                           ▼
              ┌────────────────────────┐
              │     Worker Pool (N)    │
              │  - count per category  │
              │  - count per cat+date  │
              │  local maps per worker │
              └───────────┬────────────┘
                          │ resultCh (*Result)
                          ▼
              ┌────────────────────────┐
              │     Aggregator         │
              │  merge all worker maps │
              │  find max category     │
              │  find peak date        │
              └───────────┬────────────┘
                          │
                          ▼
              ┌────────────────────────┐
              │   MD Report Writer     │
              │   → output/report.md   │
              └────────────────────────┘
```

### Design Patterns Used

| Pattern | Implementation |
|---------|---------------|
| **Fan-out/Fan-in** | Reader goroutines (fan-out) → job channel → Worker Pool (fan-in) → result channel → Aggregator |
| **Worker Pool** | N worker goroutines read from a shared channel, each with their own local maps (zero lock contention) |
| **Pipeline** | Data flows: Reader → Channel → Worker → Channel → Aggregator → Reporter |
| **Graceful Shutdown** | `context.WithCancel` — all goroutines listen on `ctx.Done()` |
| **Atomic Counters** | `atomic.Int64` for progress tracking (lock-free, concurrent-safe) |

---

## Project Structure

```
go_project/
├── main.go                        # Entry point & orchestrator
├── go.mod
├── config/
│   └── config.go                  # Configuration via environment variables
├── internal/
│   ├── csvreader/
│   │   ├── reader.go              # Concurrent CSV reader (goroutine per file)
│   │   ├── reader_test.go
│   │   ├── generator.go           # Dummy CSV generator (multi-goroutine writer)
│   │   └── generator_test.go
│   ├── workerpool/
│   │   ├── pool.go                # Worker pool (fan-out/fan-in) + worker function
│   │   ├── pool_test.go
│   │   ├── worker.go              # Worker logic (per-record processing)
│   │   ├── worker_test.go
│   │   ├── aggregator.go          # Merge results & find max category/date
│   │   └── aggregator_test.go
│   ├── tracker/
│   │   ├── progress.go            # Atomic counters + periodic report
│   │   └── progress_test.go
│   ├── model/
│   │   ├── record.go              # Data model (Record struct + CSV parser)
│   │   └── record_test.go
│   └── reporter/
│       ├── mdwriter.go            # Markdown report writer
│       └── mdwriter_test.go
├── data/                          # Input CSV directory (auto-generated if empty)
├── output/                        # Output report directory
└── vibing/                        # Planning and documentation
    ├── plan_concurrent_data.md
    └── unit_test.md
```

---

## Data Model

The CSV contains 7 columns: `id`, `name`, `email`, `amount`, `category`, `date`, `status`

```go
type Record struct {
    ID       int      // id
    Name     string   // name
    Email    string   // email
    Amount   float64  // amount
    Category string   // category (Electronics, Fashion, Food, etc.)
    Date     string   // date (YYYY-MM-DD)
    Status   string   // status (active, inactive, pending, cancelled)
}
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- Zero external dependencies (standard library only)

### 1. Navigate to project directory

```bash
cd go_project
```

### 2. Run the application

```bash
go run main.go
```

The app will:
1. Check `data/` — if no CSV files found, generate 3 dummy files (~100k rows total)
2. Read all CSV files in parallel
3. Worker pool processes the data (count category + category-date combos)
4. Aggregator finds the most common category and its peak date
5. Display real-time progress in the terminal
6. Write the report to `output/report.md`

### 3. Sample output

```
🚀 Starting Concurrent Data Processing...
Generating dummy CSV files in ./data...
Reading CSV files...
[Progress] Read: 24521 | Processed: 24512 | Errors: 0
[Progress] Read: 58342 | Processed: 58330 | Errors: 0
[Progress] Read: 100002 | Processed: 100002 | Errors: 0

✅ Processing Completed in 823ms!
Read: 100002 | Processed: 100002 | Errors: 0
📄 Report saved to output/report.md
```

### 4. Sample `output/report.md`

```markdown
# CSV Processing Report

**Generated:** 2026-08-07 14:30:00

## Summary
- Total files processed: 3
- Total rows (read attempt): 100002
- Rows processed: 100000
- Errors skipped: 2
- Duration: 823ms

## Analysis
- **Most common category:** `Electronics` (10,045 records)
- **Peak date for Electronics:** `2025-06-15` (123 records)

## Processing Details
- Workers: 8
- Throughput: 121507 rows/s
```

---

## Configuration (Environment Variables)

| Variable | Default | Description |
|----------|---------|-------------|
| `WORKERS` | `8` | Number of worker goroutines |
| `CHANNEL_BUF` | `10000` | Job channel buffer size |
| `CSV_DIR` | `./data` | Input CSV directory |
| `OUTPUT_DIR` | `./output` | Output report directory |
| `REPORT_INTERVAL` | `1s` | Progress report interval |
| `CSV_FILES` | `3` | Number of dummy files to generate |
| `CSV_ROWS` | `100000` | Total dummy rows to generate |

Example with custom config:

```bash
WORKERS=16 CSV_ROWS=500000 go run main.go
```

---

## Error Handling

| Error Type | Handling |
|------------|----------|
| **File not found** | Log error, skip file, continue with remaining files |
| **CSV parse error (single row)** | Skip row, increment error counter, continue |
| **Worker panic** | `recover()` → send error to error channel, worker continues |
| **Context cancelled** | All goroutines graceful shutdown via `ctx.Done()` |
| **Empty file** | Log warning, skip |

---

## Memory Management

| Strategy | Implementation |
|----------|---------------|
| **Buffered channels** | `CHANNEL_BUF=10000` prevents reader blocking when workers are slow |
| **Worker-local maps** | Each worker holds its own `map[string]int` — **zero lock contention**, merged later by the aggregator |
| **Channel-based pipeline** | Record lifetime is short: produced → channel → consumed, GC collects quickly |
| **Batch flush CSV generator** | `w.Flush()` every 5000 rows during generation — never holds entire dataset in memory |
| **Goroutine lifecycle** | All goroutines bound by `sync.WaitGroup`, no goroutine leaks |

---

## Running Tests

```bash
# All unit tests with race detector
go test -race -v ./internal/...

# CSV reader throughput benchmark
go test -bench=. -benchmem ./internal/csvreader/...

# Coverage report
go test -race -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Targets

| Package | Target |
|---------|--------|
| `csvreader` | 85% |
| `workerpool` | 85% |
| `tracker` | 90% |
| `model` | 95% |
| `reporter` | 80% |

---

## Detailed Flow

### Step 1: CSV Generation (if needed)
`Generator.Generate()` — multi-goroutine writer, 3 files × ~33,334 rows = ~100k rows with random realistic data.

### Step 2: CSV Reading
`Reader.ReadFiles()` — 1 goroutine per file, reads line by line using `encoding/csv`, parses into `Record` struct, sends to `jobCh` (buffered channel).

### Step 3: Worker Processing
`StartPool()` — N worker goroutines read from `jobCh`. Each worker:
- Maintains `map[string]int` for category counts
- Maintains `map[string]map[string]int` for category-date counts
- Sends `*Result` to `resultCh`

### Step 4: Aggregation
`Aggregate()` — single goroutine receives all `*Result` from `resultCh`, merges all maps, finds the category with the highest count and its peak date.

### Step 5: Report
`WriteReport()` — writes all statistics and analysis results to `output/report.md` in Markdown format.

---

## Tech Stack

- **Go 1.21+** — standard library only
- `sync.WaitGroup` — goroutine synchronization
- `sync/atomic` — lock-free atomic counters
- `encoding/csv` — CSV parsing
- `context` — cancellation propagation
- `time.Ticker` — periodic progress reporting
