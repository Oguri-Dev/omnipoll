package poller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/omnipoll/backend/internal/akva"
	"github.com/omnipoll/backend/internal/config"
	"github.com/omnipoll/backend/internal/events"
	"github.com/omnipoll/backend/internal/mongo"
	"github.com/omnipoll/backend/internal/mqtt"
)

// Worker manages the polling goroutine lifecycle
type Worker struct {
	mu            sync.RWMutex
	running       bool
	stopChan      chan struct{}
	configManager *config.Manager
	poller        *Poller
	watermark     *WatermarkManager
	akvaClient    *akva.Client
	mqttClient    *mqtt.Client
	mqttPub       *mqtt.Publisher
	mongoClient   *mongo.Client
	mongoRepo     *mongo.Repository
	logsMu        sync.Mutex
	logs          []events.LogEntry
	maxLogs       int
}

// NewWorker creates a new polling worker
func NewWorker(cfgManager *config.Manager) *Worker {
	return &Worker{
		configManager: cfgManager,
		watermark:     NewWatermarkManager(),
		maxLogs:       1000,
		logs:          make([]events.LogEntry, 0),
	}
}

// Initialize sets up all connections
func (w *Worker) Initialize(ctx context.Context) error {
	cfg := w.configManager.Get()

	// Load watermark
	if err := w.watermark.Load(); err != nil {
		w.logEntry("error", "Failed to load watermark: "+err.Error())
		return err
	}
	w.logEntry("info", "Watermark loaded")

	// Initialize Akva client
	w.akvaClient = akva.NewClient(cfg.SQLServer)
	if err := w.akvaClient.Connect(ctx); err != nil {
		w.logEntry("warn", "Failed to connect to SQL Server: "+err.Error())
		// Don't fail - worker can try to reconnect later
	} else {
		w.logEntry("info", "Connected to SQL Server")
	}

	// Initialize MQTT client
	w.mqttClient = mqtt.NewClient(cfg.MQTT)
	if err := w.mqttClient.Connect(); err != nil {
		w.logEntry("warn", "Failed to connect to MQTT: "+err.Error())
	} else {
		w.logEntry("info", "Connected to MQTT broker")
	}
	w.mqttPub = mqtt.NewPublisher(w.mqttClient)

	// Initialize MongoDB client
	w.mongoClient = mongo.NewClient(cfg.MongoDB)
	if err := w.mongoClient.Connect(ctx); err != nil {
		w.logEntry("warn", "Failed to connect to MongoDB: "+err.Error())
	} else {
		w.logEntry("info", "Connected to MongoDB")
	}
	w.mongoRepo = mongo.NewRepository(w.mongoClient)

	// Create poller
	w.poller = NewPoller(cfg.Polling, w.akvaClient, w.mqttPub, w.mongoRepo, w.watermark)

	// Refresh stats from MongoDB (only if connected)
	if w.mongoClient != nil && w.mongoClient.IsConnected() {
		w.poller.RefreshStats(ctx)
	}

	return nil
}

// Start starts the polling worker
func (w *Worker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("worker already running")
	}

	// Initialize if needed
	if w.poller == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := w.Initialize(ctx); err != nil {
			return err
		}
	}

	w.stopChan = make(chan struct{})
	w.running = true

	go w.run()

	w.logEntry("info", "Worker started")
	return nil
}

// Stop stops the polling worker
func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopChan)
	w.running = false
	w.logEntry("info", "Worker stopped")
}

// run is the main polling loop
func (w *Worker) run() {
	cfg := w.configManager.Get()
	interval := time.Duration(cfg.Polling.IntervalMS) * time.Millisecond

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	w.doPoll()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.doPoll()
		}
	}
}

// doPoll executes a single poll cycle
func (w *Worker) doPoll() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := w.poller.Poll(ctx); err != nil {
		w.logEntry("error", "Poll error: "+err.Error())
	}
}

// IsRunning returns whether the worker is running
func (w *Worker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// GetStats returns worker statistics
func (w *Worker) GetStats() Stats {
	log.Println("[DEBUG] Worker.GetStats: START")
	if w.poller == nil {
		log.Println("[DEBUG] Worker.GetStats: poller is nil, returning empty")
		return Stats{}
	}
	log.Println("[DEBUG] Worker.GetStats: Calling poller.GetStats")
	stats := w.poller.GetStats()
	log.Println("[DEBUG] Worker.GetStats: poller.GetStats returned")
	return stats
}

// ResetWatermark resets the watermark
func (w *Worker) ResetWatermark() error {
	if w.watermark == nil {
		w.watermark = NewWatermarkManager()
	}
	return w.watermark.Reset()
}

// TestSQLConnection tests SQL Server connection
func (w *Worker) TestSQLConnection() (bool, error) {
	cfg := w.configManager.Get()
	client := akva.NewClient(cfg.SQLServer)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.TestConnection(ctx)
	if err != nil {
		return false, err
	}
	client.Close()
	return true, nil
}

// TestMQTTConnection tests MQTT connection
func (w *Worker) TestMQTTConnection() (bool, error) {
	cfg := w.configManager.Get()
	client := mqtt.NewClient(cfg.MQTT)
	err := client.TestConnection()
	if err != nil {
		return false, err
	}
	client.Disconnect()
	return true, nil
}

// TestMongoConnection tests MongoDB connection
func (w *Worker) TestMongoConnection() (bool, error) {
	cfg := w.configManager.Get()
	client := mongo.NewClient(cfg.MongoDB)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.TestConnection(ctx)
	if err != nil {
		return false, err
	}
	client.Disconnect(ctx)
	return true, nil
}

// GetLogs returns recent log entries
func (w *Worker) GetLogs() []events.LogEntry {
	w.logsMu.Lock()
	defer w.logsMu.Unlock()
	return w.logs
}

// GetRecentEvents returns recent events from MongoDB
func (w *Worker) GetRecentEvents(ctx context.Context, limit int) ([]mongo.HistoricalEvent, error) {
	if w.mongoRepo == nil {
		return []mongo.HistoricalEvent{}, nil
	}
	return w.mongoRepo.GetRecentEvents(ctx, limit)
}

// logEntry adds a log entry
func (w *Worker) logEntry(level, message string) {
	entry := events.LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
	}

	log.Printf("[%s] %s", level, message)

	w.logsMu.Lock()
	w.logs = append(w.logs, entry)
	// Keep only last maxLogs entries
	if len(w.logs) > w.maxLogs {
		w.logs = w.logs[len(w.logs)-w.maxLogs:]
	}
	w.logsMu.Unlock()
}

// Shutdown gracefully shuts down the worker
func (w *Worker) Shutdown(ctx context.Context) {
	w.Stop()

	if w.akvaClient != nil {
		w.akvaClient.Close()
	}
	if w.mqttClient != nil {
		w.mqttClient.Disconnect()
	}
	if w.mongoClient != nil {
		w.mongoClient.Disconnect(ctx)
	}
}
