    # Spesifikasi Teknis: Modular Go Project Template

    ## 1. Konteks Fitur
    Transformasi repository Go menjadi *reusable template* dengan kapabilitas *Feature Toggling* berbasis environment. Memungkinkan inisiasi project instan dan skalabilitas *pay-as-you-go* secara arsitektural
  (fitur berat seperti gRPC/MQ dapat dimatikan hingga dibutuhkan).

    ## 2. Implementasi Inisiasi Project (Makefile)
    Pembuatan instruksi *find-and-replace* untuk memutasi nama *module* dasar.

    **File:** `Makefile`
    ```makefile
    # Module template bawaan
    TEMPLATE_MODULE := github.com/Rivaldz/my-template-go

    # Perintah: make init-project MODULE=github.com/nama/project-baru
    init-project:
        @if [ -z "$(MODULE)" ]; then echo "Error: MODULE is required."; exit 1; fi
        @echo "Mengubah module menjadi $(MODULE)..."
        @find . -type f \( -name '*.go' -o -name 'go.mod' -o -name 'Makefile' -o -name 'Dockerfile' \) -exec sed -i 's|


  ⎛TEMPLATE ODULE⎞|
  ⎝        M     ⎠


    (MODULE)|g' {} +
        @go mod tidy
        @echo "Project $(MODULE) siap digunakan!"

  ## 3. Implementasi Feature Toggling (Arsitektur Config & Bootstrap)

  Memanfaatkan Environment Variables untuk menyalakan/mematikan service di  main.go  tanpa harus melakukan refactoring kode atau mengubah argumen kompilasi build tags.

  File:  .env.example

    # SERVER TOGGLES
    ENABLE_REST=true
    ENABLE_GRPC=false

    # BROKER TOGGLES
    ENABLE_RABBITMQ=false
    ENABLE_NATS=false

  File:  config/config.go

    type FeatureToggles struct {
        EnableREST     bool `env:"ENABLE_REST" envDefault:"true"`
        EnableGRPC     bool `env:"ENABLE_GRPC" envDefault:"false"`
        EnableRabbitMQ bool `env:"ENABLE_RABBITMQ" envDefault:"false"`
        EnableNATS     bool `env:"ENABLE_NATS" envDefault:"false"`
    }

    type Config struct {
        // ... konfigurasi DB, App, dll
        Features FeatureToggles
    }

  File:  cmd/main.go  (Logika Bisnis Graceful Start/Shutdown)

    package main

    import (
        "context"
        "log"
        "os/signal"
        "syscall"
    )

    func main() {
        cfg := config.Load()
        ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
        defer stop()

        // Init Core Dependencies
        db := database.New(cfg)
        myUseCase := usecase.New(db)

        errChan := make(chan error, 1)

        // 1. REST Server
        if cfg.Features.EnableREST {
            restServer := rest.NewServer(cfg, myUseCase)                                                                                                                                                           
            go func() {                                                                                                                                                                                            
                log.Println("Starting REST Server...")                                                                                                                                                             
                if err := restServer.Start(); err != nil { errChan <- err }
            }()
        }

        // 2. gRPC Server
        if cfg.Features.EnableGRPC {
            grpcServer := grpc.NewServer(cfg, myUseCase)
            go func() {
                log.Println("Starting gRPC Server...")
                if err := grpcServer.Start(); err != nil { errChan <- err }
            }()
        }

        // 3. RabbitMQ Consumer
        if cfg.Features.EnableRabbitMQ {
            mqConsumer := rabbitmq.NewConsumer(cfg, myUseCase)
            go func() {
                log.Println("Starting RabbitMQ Consumer...")
                if err := mqConsumer.Consume(ctx); err != nil { errChan <- err }
            }()
        }

        // Graceful Shutdown
        select {
        case err := <-errChan:
            log.Fatalf("Critical error: %v", err)
        case <-ctx.Done():
            log.Println("Shutting down gracefully...")
            // Panggil penutup resources di sini (contoh: restServer.Shutdown())
        }
    }

  ## 4. Evaluasi Sistem (YAGNI & Kompromi Arsitektur)

  Pendekatan Config-Driven Feature Toggle ini adalah jalan tengah (pragmatic compromise) yang paling aman:

  • Mengurangi kebingungan saat onboarding (tidak perlu belajar cara merangkai gRPC jika saat ini hanya butuh REST).
  • Meskipun binary hasil  go build  mungkin lebih besar 2-3MB karena dependensi gRPC/RabbitMQ tetap ikut dikompilasi, dampaknya pada performa Runtime (RAM/CPU) adalah nol (karena goroutine listener-nya tidak
  pernah di-start). Ini adalah trade-off terbaik untuk arsitektur template level Enterprise.
