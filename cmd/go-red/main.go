package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/go-red/internal/config"
	"github.com/yourusername/go-red/internal/engine"
	"github.com/yourusername/go-red/internal/registry"
	"github.com/yourusername/go-red/internal/server"
	"github.com/yourusername/go-red/internal/storage"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to config file")
	httpPort := flag.Int("port", 1880, "HTTP port to listen on")
	flowDir := flag.String("flows", "./flows", "Directory to store flows")
	flag.Parse()

	// Initialize configuration
	cfg := config.New()
	if *configFile != "" {
		if err := cfg.LoadFromFile(*configFile); err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}
	cfg.SetDefault("http.port", *httpPort)
	cfg.SetDefault("storage.dir", *flowDir)

	// Create storage
	store, err := storage.NewFileStorage(cfg.GetString("storage.dir"))
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize node registry
	reg := registry.New()
	if err := reg.LoadBuiltinNodes(); err != nil {
		log.Fatalf("Failed to load builtin nodes: %v", err)
	}

	// Create and initialize engine
	eng := engine.New(reg, store)
	if err := eng.Initialize(); err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// Start the engine
	if err := eng.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}
	defer eng.Stop()

	// Create and start HTTP server
	srv := server.New(cfg, eng, store)
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	fmt.Printf("go-red started on port %d\n", cfg.GetInt("http.port"))
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Shutting down...")
}
