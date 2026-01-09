package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omnipoll/backend/internal/admin"
	"github.com/omnipoll/backend/internal/config"
	"github.com/omnipoll/backend/internal/poller"
)

//go:embed ../../web/dist/*
var staticFiles embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Omnipoll...")

	// Initialize configuration manager
	cfgManager, err := config.NewManager()
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}

	// Load configuration
	if err := cfgManager.Load(); err != nil {
		log.Printf("Warning: Could not load config file, using defaults: %v", err)
	}

	cfg := cfgManager.Get()
	log.Printf("Configuration loaded from: %s", cfgManager.GetPath())

	// Initialize worker
	worker := poller.NewWorker(cfgManager)

	// Initialize worker connections in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := worker.Initialize(ctx); err != nil {
			log.Printf("Worker initialization warning: %v", err)
		}
	}()

	// Create admin server
	adminServer, err := admin.NewServer(cfgManager, worker, staticFiles)
	if err != nil {
		log.Fatalf("Failed to create admin server: %v", err)
	}

	// Start admin server in goroutine
	go func() {
		log.Printf("Admin panel available at http://%s:%d", cfg.Admin.Host, cfg.Admin.Port)
		if err := adminServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Admin server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop worker
	worker.Shutdown(ctx)

	// Stop admin server
	if err := adminServer.Stop(ctx); err != nil {
		log.Printf("Error stopping admin server: %v", err)
	}

	log.Println("Omnipoll shutdown complete")
}
