.PHONY: run test build clean

# Menjalankan kode aplikasi
run:
	go run main.go

# Melakukan testing kode beserta race detector
test:
	go test -race -v ./...

# Melakukan kompilasi ke bentuk binary
build:
	go build -o app main.go

# Membersihkan file binary dan folder data/output (jika diperlukan)
clean:
	rm -f app
	rm -rf data/*.csv
	rm -rf output/*.md
