# QA Handoff Documentation: Modular Go Project Template & Feature Toggling

**Tanggal**: 2026-07-10
**Referensi**: `./vibing/analis/2026-07-09/core-project-plan.md`

## 1. Fitur
Transformasi repository Go menjadi reusable template dengan kapabilitas Feature Toggling berbasis environment variables.
Memungkinkan inisiasi project baru melalui command `make init-project MODULE=...` dan mematikan/menyalakan service (REST, gRPC, RabbitMQ, NATS) melalui konfigurasi environment.

## 2. File yang diubah/dibuat
- `Makefile`: Menambahkan rule `init-project` untuk inisiasi modular name find-and-replace, dan `go mod tidy`.
- `.env.example`: Menambahkan SERVER TOGGLES (`ENABLE_REST`, `ENABLE_GRPC`) dan BROKER TOGGLES (`ENABLE_RABBITMQ`, `ENABLE_NATS`).
- `config/config.go`: Menambahkan struct `FeatureToggles` untuk menangkap nilai environment variables.
- `internal/app/app.go`: Mengubah logika inisiasi dan graceful shutdown agar berjalan secara kondisional sesuai dengan nilai environment toggle yang aktif (tidak menjalankan server/listener yang di-disable).

## 3. Kontrak API
*Tidak ada penambahan/perubahan kontrak API HTTP.* Fitur ini mengatur infrastruktur layer dan server instantiation.

## 4. Skema Database
*Tidak ada perubahan skema database atau migration.*

## 5. Testing
**Command**: `make deps && make test`
**Summary**:
```text
PASS
ok  	github.com/Rivaldz/my-template-go/internal/usecase	2.774s
PASS
ok  	github.com/Rivaldz/my-template-go/pkg/jwt	1.038s
PASS
ok  	github.com/Rivaldz/my-template-go/pkg/logger	1.056s
```
Seluruh komponen service berhasil lulus kompilasi, dependensi verified, dan suite pengujian utama dinyatakan sukses.

## 6. Cara Verifikasi
1. **Verifikasi Feature Toggling**:
   - Jalankan `make run` dengan konfigurasi `.env` standar (semua toggle enabled/disabled).
   - Matikan RabbitMQ dan gRPC via `.env` (`ENABLE_RABBITMQ=false`, `ENABLE_GRPC=false`).
   - Perhatikan pada log console, proses yang tidak di-enable tidak akan memulai server.
   - Uji graceful shutdown dengan mengirim *SIGINT* (Ctrl+C). Pastikan service yang non-aktif tidak memicu error channel shutdown.

2. **Verifikasi init-project**:
   - Lakukan commit sementara atau copy repository.
   - Jalankan `make init-project MODULE=github.com/dummy/my-app`.
   - Pastikan seluruh package import di file `*.go`, isi `go.mod`, dan `Makefile` berubah menjadi module yang baru secara konsisten.
   - Pastikan aplikasi masih dapat di-build (`go build ./cmd/app`) setelah diganti namanya.

## 7. Catatan
- Pendekatan toggling tidak mengubah file build target. Binary masih mencakup dependensi (gRPC/MQ), namun tidak menghabiskan runtime resources (goroutine) jika tidak diswitch menyala.
- Implementasi sesuai spesifikasi yagni analyst tanpa technical debt terdeteksi.
