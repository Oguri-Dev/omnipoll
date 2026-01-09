package poller

import (
	"context"
	"log"
	"time"

	"github.com/omnipoll/backend/internal/akva"
	"github.com/omnipoll/backend/internal/config"
	"github.com/omnipoll/backend/internal/mongo"
	"github.com/omnipoll/backend/internal/mqtt"
)

// Poller handles incremental data extraction with watermark management
type Poller struct {
	config     config.PollingConfig
	akvaClient *akva.Client
	mqttPub    *mqtt.Publisher
	mongoRepo  *mongo.Repository
	watermark  *WatermarkManager
	stats      *Stats
}

// Stats tracks polling statistics
type Stats struct {
	LastFechaHora  time.Time
	EventsToday    int64
	TotalEvents    int64
	IngestionRate  float64
	SQLConnected   bool
	MQTTConnected  bool
	MongoConnected bool

	// For rate calculation
	lastMinuteEvents int64
	lastRateCalc     time.Time
}

// NewPoller creates a new poller instance
func NewPoller(
	cfg config.PollingConfig,
	akvaClient *akva.Client,
	mqttPub *mqtt.Publisher,
	mongoRepo *mongo.Repository,
	watermark *WatermarkManager,
) *Poller {
	return &Poller{
		config:     cfg,
		akvaClient: akvaClient,
		mqttPub:    mqttPub,
		mongoRepo:  mongoRepo,
		watermark:  watermark,
		stats: &Stats{
			lastRateCalc: time.Now(),
		},
	}
}

// Poll executes one polling cycle
func (p *Poller) Poll(ctx context.Context) error {
	// Get current watermark
	wm := p.watermark.Get()

	// Fetch new records from Akva
	records, err := p.akvaClient.FetchNewRecords(ctx, wm.LastFechaHora, wm.IDsAtLastFechaHora, p.config.BatchSize)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil // No new records
	}

	log.Printf("Fetched %d new records from Akva", len(records))

	// Convert to normalized events
	events := akva.ToNormalizedEvents(records)

	// Publish to MQTT
	if err := p.mqttPub.PublishBatch(events); err != nil {
		return err
	}

	log.Printf("Published %d events to MQTT", len(events))

	// Persist to MongoDB
	if err := p.mongoRepo.InsertBatch(ctx, events); err != nil {
		log.Printf("Warning: MongoDB insert error (may be duplicates): %v", err)
		// Continue anyway - duplicates are expected for idempotency
	}

	log.Printf("Persisted %d events to MongoDB", len(events))

	// Update watermark
	// Find the latest timestamp and collect IDs at that timestamp
	var latestTime time.Time
	var idsAtLatest []string

	for _, record := range records {
		if record.FechaHora.After(latestTime) {
			latestTime = record.FechaHora
			idsAtLatest = []string{record.ID}
		} else if record.FechaHora.Equal(latestTime) {
			idsAtLatest = append(idsAtLatest, record.ID)
		}
	}

	if err := p.watermark.Update(latestTime, idsAtLatest); err != nil {
		return err
	}

	// Update stats
	p.updateStats(latestTime, int64(len(records)))

	return nil
}

// updateStats updates polling statistics
func (p *Poller) updateStats(lastFechaHora time.Time, newEvents int64) {
	p.stats.LastFechaHora = lastFechaHora
	p.stats.EventsToday += newEvents
	p.stats.TotalEvents += newEvents
	p.stats.lastMinuteEvents += newEvents

	// Calculate rate every minute
	if time.Since(p.stats.lastRateCalc) >= time.Minute {
		p.stats.IngestionRate = float64(p.stats.lastMinuteEvents)
		p.stats.lastMinuteEvents = 0
		p.stats.lastRateCalc = time.Now()
	}
}

// GetStats returns current statistics
func (p *Poller) GetStats() Stats {
	stats := *p.stats
	stats.SQLConnected = p.akvaClient.IsConnected()
	stats.MQTTConnected = p.mqttPub.IsConnected()
	stats.MongoConnected = p.mongoRepo.IsConnected()
	return stats
}

// RefreshStats refreshes statistics from MongoDB
func (p *Poller) RefreshStats(ctx context.Context) {
	if total, err := p.mongoRepo.CountEvents(ctx); err == nil {
		p.stats.TotalEvents = total
	}
	if today, err := p.mongoRepo.CountEventsToday(ctx); err == nil {
		p.stats.EventsToday = today
	}
}
