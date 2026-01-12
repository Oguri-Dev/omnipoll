package main

import (
	"context"
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

	// Initialize and auto-start worker in background
	go func() {
		// Initialize with reasonable timeout (connections will retry later)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Initialize (ignore errors - connections will be retried during polling)
		_ = worker.Initialize(ctx)
		
		// Start worker immediately after init attempt
		if err := worker.Start(); err != nil {
			log.Printf("Failed to start worker: %v", err)
		} else {
			log.Println("✓ Worker started automatically")
		}
	}()

	// Create admin server (files served from filesystem at /app/web/dist in Docker)
	adminServer := admin.NewServerWithFilesystem(cfgManager, worker, "./web/dist")

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
