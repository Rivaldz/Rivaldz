# Plan Unit Test - Concurrent Data Processing

## Coverage Target

```
internal/
├── csvreader/
│   ├── reader.go          → reader_test.go
│   └── generator.go       → generator_test.go
├── workerpool/
│   ├── pool.go            → pool_test.go
│   ├── worker.go          → worker_test.go
│   └── aggregator.go      → aggregator_test.go
├── tracker/
│   └── progress.go        → progress_test.go
├── model/
│   └── record.go          → record_test.go
└── reporter/
    └── mdwriter.go        → mdwriter_test.go
```

## Testing Strategy

### 1. Approach
- **Table-driven tests** (`t.Run` subtests) — standar Go
- **Parallel tests** (`t.Parallel()`) untuk concurrency correctness
- Semua test dijalankan dengan **race detector**: `go test -race ./...`
- **Benchmark tests** untuk CSV reader dan worker pool throughput
- **Integration test** kecil untuk end-to-end flow

### 2. Test Fixtures
- Gunakan `t.TempDir()` untuk file temporer (otomatis cleanup)
- Embed CSV kecil sebagai string constant untuk test parsing
- Mock generator menghasilkan dataset kecil (100-1000 rows)

---

## Unit Test Cases

### `internal/csvreader/reader_test.go`

| Test Name                         | What it tests                                           |
|-----------------------------------|---------------------------------------------------------|
| `TestReadCSV_SingleFile`          | Baca 1 file CSV valid, cek semua record terkirim        |
| `TestReadCSV_MultipleFiles`       | Baca 3 file paralel, cek semua record masuk channel     |
| `TestReadCSV_EmptyFile`           | File kosong → tidak ada record, tidak crash             |
| `TestReadCSV_MalformedRows`       | Beberapa baris corrupt → skip, kirim error ke errCh     |
| `TestReadCSV_FileNotFound`        | File tidak ada → error di errCh, reader lanjut          |
| `TestReadCSV_ContextCancellation` | `ctx.Done()` → reader graceful shutdown                 |
| `TestReadCSV_LargeFile`           | File 10k rows → tidak ada memory leak / deadlock        |
| `BenchmarkReadCSV`                | Benchmark throughput (rows/sec)                         |

### `internal/csvreader/generator_test.go`

| Test Name                    | What it tests                                        |
|------------------------------|------------------------------------------------------|
| `TestGenerateCSV_RowCount`   | Generate 100 rows → cek jumlah baris di file         |
| `TestGenerateCSV_Columns`    | Cek jumlah kolom = 7, header valid                   |
| `TestGenerateCSV_DataTypes`  | ID numeric, amount float, date format YYYY-MM-DD     |
| `TestGenerateCSV_MultipleFiles` | Generate 3 files → semua ada, semua punya header  |

### `internal/model/record_test.go`

| Test Name                    | What it tests                                        |
|------------------------------|------------------------------------------------------|
| `TestParseRecord_Valid`      | CSV row valid → struct terisi benar                  |
| `TestParseRecord_InvalidColumns` | Jumlah kolom salah → error                       |
| `TestParseRecord_BadInt`     | ID bukan angka → error                               |
| `TestParseRecord_BadFloat`   | Amount bukan angka → error                           |
| `TestParseRecord_EmptyFields` | Fields kosong → default / error sesuai               |

### `internal/workerpool/worker_test.go`

| Test Name                    | What it tests                                        |
|------------------------------|------------------------------------------------------|
| `TestWorker_CountCategory`   | Kirim 5 record → cek category map di aggregator      |
| `TestWorker_CountCategoryDate` | Kirim record beda date → cek category-date map    |
| `TestWorker_PanicRecovery`   | Record invalid trigger panic → worker recover, lanjut|
| `TestWorker_StopChannel`     | Job channel close → worker exit clean                |
| `TestWorker_ContextDone`     | ctx cancel → worker exit via ctx.Done()              |

### `internal/workerpool/pool_test.go`

| Test Name                    | What it tests                                        |
|------------------------------|------------------------------------------------------|
| `TestPool_ProcessRecords`    | 1000 record → workers proses, cek aggregator results |
| `TestPool_ConcurrentSafety`  | Banyak worker akses channel → no race, hasil konsisten|
| `TestPool_ErrorHandling`     | Ada record corrupt → error channel terisi, not crash |
| `TestPool_ZeroWorkers`       | Workers=0 → fallback ke 1 / error                     |
| `TestPool_GracefulShutdown`  | ctx.Cancel mid-process → semua worker clean exit      |

### `internal/workerpool/aggregator_test.go`

| Test Name                       | What it tests                                     |
|---------------------------------|---------------------------------------------------|
| `TestAggregate_MergeMaps`       | 2 worker maps digabung → total count benar        |
| `TestAggregate_FindMax`         | Dari merged map → category max & date puncak      |
| `TestAggregate_SingleCategory`  | Hanya 1 category → result benar                   |
| `TestAggregate_EmptyInput`      | Map kosong → result "empty" / zero value          |

### `internal/tracker/progress_test.go`

| Test Name                       | What it tests                                     |
|---------------------------------|---------------------------------------------------|
| `TestProgress_IncrementTotal`   | `IncrementTotal(N)` → `Total()` == N              |
| `TestProgress_IncrementErrors`  | `IncrementErrors(N)` → `Errors()` == N            |
| `TestProgress_ConcurrentInc`    | 100 goroutine increment → atomic, no data race    |
| `TestProgress_ReportInterval`   | Progress reported tiap interval                   |
| `TestProgress_Summary`          | Stop() → return summary struct yang valid          |

### `internal/reporter/mdwriter_test.go`

| Test Name                       | What it tests                                     |
|---------------------------------|---------------------------------------------------|
| `TestMDWriter_BasicOutput`      | Write summary → file ada, format markdown benar   |
| `TestMDWriter_FilePermissions`  | Cek file readable setelah write                   |
| `TestMDWriter_Overwrite`        | Write dua kali → file ter-overwrite               |
| `TestMDWriter_EmptyData`        | Summary dengan 0 data → tetap ter-render valid    |

---

## Integration Test (optional, `main_test.go` atau internal `e2e_test.go`)

| Test Name                       | What it tests                                     |
|---------------------------------|---------------------------------------------------|
| `TestE2E_SmallDataset`          | Generate 100 rows 3 file → full pipeline → cek .md |
| `TestE2E_NoFiles`               | Folder kosong → tidak crash, report "no data"      |

---

## Menjalankan Test

```bash
# Unit test with race detector
go test -race -v ./internal/...

# Benchmark
go test -bench=. -benchmem ./internal/csvreader/...

# Coverage report
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html

# Run all + race + coverage
go test -race -coverprofile=coverage.out ./...
```

## Target Coverage Minimum

| Package     | Coverage Target |
|-------------|-----------------|
| csvreader   | 85%             |
| workerpool  | 85%             |
| tracker     | 90%             |
| model       | 95%             |
| reporter    | 80%             |
