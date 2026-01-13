package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/omnipoll/backend/internal/events"
)

// Publisher handles MQTT message publishing
type Publisher struct {
	client *Client
}

// NewPublisher creates a new MQTT publisher
func NewPublisher(client *Client) *Publisher {
	return &Publisher{
		client: client,
	}
}

// MQTTMessage represents the message format for MQTT (all fields from TB_DetalleAlimentacion)
type MQTTMessage struct {
	ID                  string  `json:"ID"`
	Centro              string  `json:"Centro"`              // Name
	Jaula               string  `json:"Jaula"`               // UnitName (cleaned)
	TimeStampAkva       string  `json:"TimeStampAkva"`       // FechaHora
	Dia                 string  `json:"Dia"`
	Inicio              string  `json:"Inicio"`
	Fin                 string  `json:"Fin"`
	Dif                 int     `json:"Dif"`
	Gramos              float64 `json:"Gramos"`              // AmountGrams
	PelletFishMin       float64 `json:"PelletFishMin"`
	Peces               float64 `json:"Peces"`               // FishCount
	PesoPromedio        float64 `json:"PesoPromedio"`        // PesoProm
	Biomasa             float64 `json:"Biomasa"`
	PelletPK            float64 `json:"PelletPK"`
	Alimento            string  `json:"Alimento"`            // Feedname
	Silo                string  `json:"Silo"`                // SiloName
	Dosificador         string  `json:"Dosificador"`         // DoserName
	GramsPorSegundo     float64 `json:"GramsPorSegundo"`     // gramspersec
	KgTonMin            float64 `json:"KgTonMin"`
	Marca               int     `json:"Marca"`
	TimeStampIngresado  string  `json:"TimeStampIngresado"`  // IngestedAt
}

// buildDynamicTopic creates topic: feeding/mowi/{centro}/
func (p *Publisher) buildDynamicTopic(centerName string) string {
	// Normalize center name: lowercase, replace spaces with underscores
	normalized := strings.ToLower(centerName)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	// Remove special characters except underscores
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	normalized = reg.ReplaceAllString(normalized, "")
	return fmt.Sprintf("feeding/mowi/%s/", normalized)
}

// cleanJaula removes letters, spaces, and special characters from unit name
func (p *Publisher) cleanJaula(unitName string) string {
	// Keep only digits
	reg := regexp.MustCompile(`[^0-9]`)
	return reg.ReplaceAllString(unitName, "")
}

// Publish publishes a single event to MQTT with dynamic topic
func (p *Publisher) Publish(event events.NormalizedEvent) error {
	client := p.client.GetClient()
	cfg := p.client.GetConfig()
	
	if client == nil || !client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	// Build dynamic topic based on center name
	topic := p.buildDynamicTopic(event.Name)
	
	// Transform to MQTT message format with ALL fields
	msg := MQTTMessage{
		ID:                 event.ID,
		Centro:             event.Name,
		Jaula:              p.cleanJaula(event.UnitName),
		TimeStampAkva:      event.FechaHora,
		Dia:                event.Dia,
		Inicio:             event.Inicio,
		Fin:                event.Fin,
		Dif:                event.Dif,
		Gramos:             event.AmountGrams,
		PelletFishMin:      event.PelletFishMin,
		Peces:              event.FishCount,
		PesoPromedio:       event.PesoProm,
		Biomasa:            event.Biomasa,
		PelletPK:           event.PelletPK,
		Alimento:           event.FeedName,
		Silo:               event.SiloName,
		Dosificador:        event.DoserName,
		GramsPorSegundo:    event.GramsPerSec,
		KgTonMin:           event.KgTonMin,
		Marca:              event.Marca,
		TimeStampIngresado: event.IngestedAt.Format(time.RFC3339),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	token := client.Publish(topic, cfg.QoS, false, payload)
	if token.WaitTimeout(2 * time.Second) && token.Error() != nil {
		return fmt.Errorf("failed to publish message: %w", token.Error())
	}

	return nil
}

// PublishBatch publishes multiple events to MQTT
func (p *Publisher) PublishBatch(evts []events.NormalizedEvent) error {
	successCount := 0
	errorCount := 0
	total := len(evts)
	cfg := p.client.GetConfig()
	
	// Log broker info once at the start
	log.Printf("[MQTT] Starting batch publish to broker: %s:%d (TLS: %v, QoS: %d)", 
		cfg.Broker, cfg.Port, cfg.UseTLS, cfg.QoS)
	
	for i, event := range evts {
		// Build topic for logging
		topic := p.buildDynamicTopic(event.Name)
		
		// Detailed logging for first 3 events only
		if i < 3 {
			log.Printf("[MQTT] Event %d/%d - Topic: %s", i+1, total, topic)
			log.Printf("[MQTT] Event %d/%d - Data: Centro=%s, Jaula=%s, Gramos=%.2f, Peces=%.0f, Biomasa=%.0f, ID=%s", 
				i+1, total, event.Name, p.cleanJaula(event.UnitName), event.AmountGrams, event.FishCount, event.Biomasa, event.ID)
		}
		
		if err := p.Publish(event); err != nil {
			errorCount++
			// Log first 5 errors in detail
			if errorCount <= 5 {
				log.Printf("[MQTT] ERROR event %d/%d (ID: %s, Topic: %s): %v", i+1, total, event.ID, topic, err)
			}
		} else {
			successCount++
		}
		
		// Log progress every 10 events
		if (i+1)%10 == 0 {
			log.Printf("[MQTT] Progress: %d/%d events (%d success, %d errors)", i+1, total, successCount, errorCount)
		}
	}
	
	log.Printf("[MQTT] Batch complete: %d/%d events published successfully (%d errors)", successCount, total, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("published %d/%d events (%d errors)", successCount, len(evts), errorCount)
	}
	
	return nil
}

// IsConnected returns whether the publisher is ready
func (p *Publisher) IsConnected() bool {
	return p.client.IsConnected()
}
