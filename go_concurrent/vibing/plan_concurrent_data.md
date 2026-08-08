# Plan Concurrent Data Processing (Go)

Proyek Go untuk memproses banyak file CSV secara bersamaan menggunakan goroutines & channels, worker pool, error handling yang graceful, dan progress tracking.

## Tujuan

- Membaca banyak file CSV secara simultan (goroutines & channels)
- Memproses data dengan pola worker pool
- Menangani error secara graceful
- Menyediakan progress tracking
- Processing logic: menampilkan kategori dengan record terbanyak beserta tanggal puncaknya
- Output: file `.md` deskriptif

## Arsitektur Proyek

```
go_project/
├── main.go                 # Entry point & orchestrator
├── go.mod
├── config/
│   └── config.go           # Konfigurasi (worker count, buffer size, dll)
├── internal/
│   ├── csvreader/
│   │   ├── reader.go       # Goroutine per file, baca baris → channel
│   │   └── generator.go    # Generate dummy CSV ~100k rows
│   ├── workerpool/
│   │   ├── pool.go         # Worker pool (fan-out/fan-in pattern)
│   │   ├── worker.go       # Worker: hitung category & category-date
│   │   └── aggregator.go   # Merge map & tentukan max
│   ├── tracker/
│   │   └── progress.go     # Atomic counter + periodic report
│   ├── model/
│   │   └── record.go       # Struct data model
│   └── reporter/
│       └── mdwriter.go     # Tulis output/report.md
├── data/                   # Folder input CSV
├── output/                 # Folder output report
└── vibing/                 # Dokumentasi plan
```

## Flow Data

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
                           │ jobCh (Record)
                           ▼
              ┌────────────────────────┐
              │     Worker Pool (N)    │
              │  - count per category  │
              │  - count per cat+date  │
              │  sync.Map per worker   │
              └───────────┬────────────┘
                          │ resultCh
                          ▼
              ┌────────────────────────┐
              │     Aggregator         │
              │  merge semua worker    │
              │  cari category max     │
              │  cari date puncaknya   │
              └───────────┬────────────┘
                          │
                          ▼
              ┌────────────────────────┐
              │   MD Report Writer     │
              │   → output/report.md   │
              └────────────────────────┘
```

## Data Model

Kolom CSV: `id, name, email, amount, category, date, status`

```go
type Record struct {
    ID       int
    Name     string
    Email    string
    Amount   float64
    Category string
    Date     string // YYYY-MM-DD
    Status   string
}
```

## Komponen Utama

### 1. CSV Reader (`internal/csvreader/reader.go`)
- Satu goroutine per file CSV
- Gunakan `encoding/csv` bawaan Go → cepat
- `bufio` reader untuk throughput tinggi
- Kirim tiap baris ke **job channel** (buffered)
- Error kirim ke **error channel** (tidak crash)
- Selesai → close channel via `WaitGroup`

### 2. Worker Pool (`internal/workerpool/`)
- N worker goroutine (configurable, default `runtime.NumCPU()`)
- Baca dari job channel, proses, kirim ke result channel
- Panic recovery per worker → tidak bikin crash
- Context cancellation → graceful shutdown
- Tiap worker punya map lokal sendiri (no lock contention)

### 3. Progress Tracker (`internal/tracker/progress.go`)
- `atomic.Int64` counters: `totalRead`, `totalProcessed`, `totalErrors`
- Ticker goroutine: log tiap interval
- Channel-based summary saat selesai

### 4. Error Handling Strategy

| Error Type                 | Handling                                   |
|----------------------------|--------------------------------------------|
| File not found             | Log, skip file, lanjut                     |
| CSV parse error (1 baris)  | Skip row, increment error counter, lanjut  |
| Worker panic               | `recover()` → log, worker lanjut           |
| Fatal (context cancelled)  | Semua goroutine graceful shutdown          |
| Empty file                 | Warn, skip                                 |

### 5. Memory Management

| Strategy              | Implementasi                                        |
|-----------------------|-----------------------------------------------------|
| `sync.Pool`           | Reuse alokasi `Record` struct                       |
| Buffered channels     | `CHANNEL_BUF=10000` biar reader ga blocking         |
| Worker-local maps     | Tiap worker punya map sendiri, merge di akhir       |
| Batch finalize        | Setelah merge, release worker maps agar GC collect  |

## Processing Logic

1. **Category Counter** → `map[string]int` — count per kategori
2. **Category-Date Counter** → `map[string]map[string]int` — count per pasangan (category, date)
3. Setelah semua worker selesai, **aggregator single goroutine** merge semua maps lalu tentukan:
   - Kategori dengan record terbanyak
   - Tanggal dengan record terbanyak untuk kategori tersebut

## Output Report (`output/report.md`)

```markdown
# CSV Processing Report

**Generated:** 2026-08-07 14:30:00

## Summary
- Total files processed: 3
- Total rows: 300,000
- Rows processed: 299,987
- Errors skipped: 13
- Duration: 1.2s

## Analysis
- **Most common category:** `Electronics` (45,231 records)
- **Peak date for Electronics:** `2026-03-15` (1,203 records)

## Processing Details
- Workers: 8
- Throughput: 250,000 rows/s
```

## Konfigurasi (via env/flag)

| Param            | Default      | Deskripsi                 |
|------------------|--------------|---------------------------|
| `WORKERS`        | `8`          | Jumlah worker goroutine   |
| `CHANNEL_BUF`    | `10000`      | Buffer channel job        |
| `CSV_DIR`        | `./data`     | Folder input CSV          |
| `REPORT_INTERVAL`| `1s`         | Interval progress report  |

## Dummy CSV Generator (`internal/csvreader/generator.go`)

- **3 file** × ~33,334 rows = total ~100k rows
- Kolom: `id, name, email, amount, category, date, status`
- Categories (random): `Electronics`, `Fashion`, `Food`, `Sports`, `Books`, `Home`, `Garden`, `Toys`, `Health`, `Automotive`
- Dates: random 2025-2026
- Status: `active`, `inactive`, `pending`, `cancelled`
- Data random namun realistis (`math/rand`)
- Multi goroutine writer supaya generation juga cepat
