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

// Poller handles incremental data extraction with watermark management
type Poller struct {
	config     config.PollingConfig
	akvaClient *akva.Client
	mqttPub    *mqtt.Publisher
	mongoRepo  *mongo.Repository
	watermark  *WatermarkManager
	stats      *Stats
	statsMu    sync.RWMutex
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
	// Update connection status before attempting operations
	p.UpdateConnectionStats()

	// Check if we have the required clients
	if p.akvaClient == nil {
		return fmt.Errorf("not connected to SQL Server (Akva)")
	}
	if p.mqttPub == nil {
		return fmt.Errorf("not connected to MQTT")
	}
	if p.mongoRepo == nil {
		return fmt.Errorf("not connected to MongoDB")
	}

	// Get current watermark
	wm := p.watermark.Get()

	// Fetch new records from Akva
	records, err := p.akvaClient.FetchNewRecords(ctx, wm.LastFechaHora, wm.IDsAtLastFechaHora, p.config.BatchSize)
	if err != nil {
		p.statsMu.Lock()
		p.stats.SQLConnected = false
		p.statsMu.Unlock()
		return err
	}

	if len(records) == 0 {
		return nil // No new records
	}

	log.Printf("Fetched %d new records from Akva", len(records))

	// Convert to normalized events
	events := akva.ToNormalizedEvents(records)

	// Filter only changed/new events for MQTT publishing
	changedEvents, err := p.filterChangedEvents(ctx, events)
	if err != nil {
		log.Printf("Warning: failed to filter changed events: %v", err)
		changedEvents = events // Fallback: publish all if filtering fails
	}

	// Publish only changed events to MQTT
	if len(changedEvents) > 0 {
		log.Printf("Attempting to publish %d changed events to MQTT", len(changedEvents))
		if err := p.mqttPub.PublishBatch(changedEvents); err != nil {
			log.Printf("ERROR: Failed to publish to MQTT: %v", err)
			return err
		}
		log.Printf("Published %d changed events to MQTT (fetched %d total)", len(changedEvents), len(events))
	} else {
		log.Printf("No changes detected (fetched %d records)", len(events))
	}

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
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	
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
	
	// Set connections as true if we're successfully polling
	p.stats.SQLConnected = true
	p.stats.MQTTConnected = true
	p.stats.MongoConnected = true
}

// GetStats returns current statistics
func (p *Poller) GetStats() Stats {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	return *p.stats
}

// RefreshStats refreshes statistics from MongoDB
func (p *Poller) RefreshStats(ctx context.Context) {
	p.statsMu.Lock()
	defer p.statsMu.Unlock()

	if p.mongoRepo == nil {
		p.stats.MongoConnected = false
	} else {
		p.stats.MongoConnected = p.mongoRepo.IsConnected()
	}

	if p.mqttPub == nil {
		p.stats.MQTTConnected = false
	} else {
		p.stats.MQTTConnected = p.mqttPub.IsConnected()
	}

	if p.akvaClient == nil {
		p.stats.SQLConnected = false
	} else {
		p.stats.SQLConnected = p.akvaClient.IsConnected()
	}
}

// UpdateConnectionStats updates connection status based on real client states
func (p *Poller) UpdateConnectionStats() {
	p.statsMu.Lock()
	defer p.statsMu.Unlock()

	if p.mqttPub != nil {
		p.stats.MQTTConnected = p.mqttPub.IsConnected()
	} else {
		p.stats.MQTTConnected = false
	}

	if p.akvaClient != nil {
		p.stats.SQLConnected = p.akvaClient.IsConnected()
	} else {
		p.stats.SQLConnected = false
	}

	if p.mongoRepo != nil {
		p.stats.MongoConnected = p.mongoRepo.IsConnected()
	} else {
		p.stats.MongoConnected = false
	}
}

// filterChangedEvents compares new events with existing MongoDB data
// and returns only those that have changes in key business fields
func (p *Poller) filterChangedEvents(ctx context.Context, newEvents []events.NormalizedEvent) ([]events.NormalizedEvent, error) {
	if len(newEvents) == 0 {
		return newEvents, nil
	}

	// Extract IDs from new events
	ids := make([]string, len(newEvents))
	for i, event := range newEvents {
		ids[i] = event.ID
	}

	// Fetch existing events from MongoDB
	existingEvents, err := p.mongoRepo.GetEventsByIDs(ctx, "akva", ids)
	if err != nil {
		return nil, err
	}

	// Compare and filter
	var changedEvents []events.NormalizedEvent
	for _, newEvent := range newEvents {
		mongoID := fmt.Sprintf("akva:%s", newEvent.ID)
		existingEvent, exists := existingEvents[mongoID]

		// If event doesn't exist in MongoDB, it's new - include it
		if !exists {
			changedEvents = append(changedEvents, newEvent)
			continue
		}

		// Compare key business fields for changes
		if hasChanges(newEvent, existingEvent) {
			changedEvents = append(changedEvents, newEvent)
		}
	}

	return changedEvents, nil
}

// hasChanges compares a new event with existing MongoDB data
// Returns true if there are meaningful changes in business fields
func hasChanges(newEvent events.NormalizedEvent, existing mongo.HistoricalEvent) bool {
	// Compare numeric business metrics
	if getFloat(existing.Payload, "amountGrams") != newEvent.AmountGrams {
		return true
	}
	if getFloat(existing.Payload, "biomasa") != newEvent.Biomasa {
		return true
	}
	if getFloat(existing.Payload, "fishCount") != newEvent.FishCount {
		return true
	}
	if getFloat(existing.Payload, "pesoProm") != newEvent.PesoProm {
		return true
	}
	if getFloat(existing.Payload, "pelletFishMin") != newEvent.PelletFishMin {
		return true
	}
	if getFloat(existing.Payload, "pelletPK") != newEvent.PelletPK {
		return true
	}
	if getFloat(existing.Payload, "gramsPerSec") != newEvent.GramsPerSec {
		return true
	}
	if getFloat(existing.Payload, "kgTonMin") != newEvent.KgTonMin {
		return true
	}

	// Compare string fields
	if getString(existing.Payload, "feedName") != newEvent.FeedName {
		return true
	}
	if getString(existing.Payload, "siloName") != newEvent.SiloName {
		return true
	}
	if getString(existing.Payload, "doserName") != newEvent.DoserName {
		return true
	}
	if getString(existing.Payload, "name") != newEvent.Name {
		return true
	}

	// Compare time fields
	if getString(existing.Payload, "inicio") != newEvent.Inicio {
		return true
	}
	if getString(existing.Payload, "fin") != newEvent.Fin {
		return true
	}

	// Compare integer fields
	if getInt(existing.Payload, "dif") != newEvent.Dif {
		return true
	}
	if getInt(existing.Payload, "marca") != newEvent.Marca {
		return true
	}

	return false
}

// Helper functions to safely extract values from payload map
func getFloat(payload map[string]interface{}, key string) float64 {
	if val, ok := payload[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

func getString(payload map[string]interface{}, key string) string {
	if val, ok := payload[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(payload map[string]interface{}, key string) int {
	if val, ok := payload[key]; ok {
		// BSON might store as int32 or int64
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}
