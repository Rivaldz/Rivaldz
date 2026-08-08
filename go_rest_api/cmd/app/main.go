package main

import (
	"log"

	"github.com/Rivaldz/my-template-go/config"
	"github.com/Rivaldz/my-template-go/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
