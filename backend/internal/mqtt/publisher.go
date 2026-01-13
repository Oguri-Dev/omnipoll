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
	
	// For QoS 0, don't wait (fire and forget)
	if cfg.QoS == 0 {
		// Just check if publish was initiated, don't wait for completion
		go func() {
			if token.Error() != nil {
				log.Printf("[MQTT] Async publish error for topic %s: %v", topic, token.Error())
			}
		}()
		return nil
	}
	
	// For QoS 1+, wait with timeout and check result
	if !token.WaitTimeout(2 * time.Second) {
		log.Printf("[MQTT] Timeout publishing to %s (payload size: %d bytes)", topic, len(payload))
		return fmt.Errorf("publish timeout after 2s for topic %s", topic)
	}
	
	if token.Error() != nil {
		return fmt.Errorf("failed to publish to %s: %w", topic, token.Error())
	}

	return nil
}

// PublishBatch publishes multiple events to MQTT
func (p *Publisher) PublishBatch(evts []events.NormalizedEvent) error {
	successCount := 0
	errorCount := 0
	total := len(evts)
	cfg := p.client.GetConfig()
	
	log.Printf("[MQTT] Publishing %d events to %s:%d (QoS: %d)", total, cfg.Broker, cfg.Port, cfg.QoS)
	
	for i, event := range evts {
		if err := p.Publish(event); err != nil {
			errorCount++
			// Log first error only
			if errorCount == 1 {
				topic := p.buildDynamicTopic(event.Name)
				log.Printf("[MQTT] First error - Topic: %s, Error: %v", topic, err)
			}
		} else {
			successCount++
		}
		
		// Log progress every 25 events
		if (i+1)%25 == 0 {
			log.Printf("[MQTT] Progress: %d/%d (%d ok, %d err)", i+1, total, successCount, errorCount)
		}
	}
	
	log.Printf("[MQTT] Complete: %d/%d published successfully", successCount, total)
	
	if errorCount > 0 {
		return fmt.Errorf("published %d/%d events (%d errors)", successCount, len(evts), errorCount)
	}
	
	return nil
}

// IsConnected returns whether the publisher is ready
func (p *Publisher) IsConnected() bool {
	return p.client.IsConnected()
}
