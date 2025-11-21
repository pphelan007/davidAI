package main

import (
	"log"

	"github.com/yourusername/gostarter/internal"
	"github.com/yourusername/gostarter/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Run the worker (blocks until shutdown)
	if err := internal.Run(cfg); err != nil {
		log.Fatalf("Worker error: %v", err)
	}
}
