package mqtt

import (
	"encoding/json"
	"fmt"
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

// MQTTMessage represents the message format for MQTT
type MQTTMessage struct {
	TimeStampAkva       string  `json:"TimeStampAkva"`
	TimeStampIngresado  string  `json:"TimeStampIngresado"`
	Jaula               string  `json:"Jaula"`
	Centro              string  `json:"Centro"`
	Gramos              float64 `json:"Gramos"`
	Biomasa             float64 `json:"Biomasa"`
	Peces               float64 `json:"Peces"`
	PesoPromedio        float64 `json:"PesoPromedio"`
	Alimento            string  `json:"Alimento"`
	Silo                string  `json:"Silo"`
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
	
	// Transform to MQTT message format
	msg := MQTTMessage{
		TimeStampAkva:      event.FechaHora,
		TimeStampIngresado: event.IngestedAt.Format(time.RFC3339),
		Jaula:              p.cleanJaula(event.UnitName),
		Centro:             event.Name,
		Gramos:             event.AmountGrams,
		Biomasa:            event.Biomasa,
		Peces:              event.FishCount,
		PesoPromedio:       event.PesoProm,
		Alimento:           event.FeedName,
		Silo:               event.SiloName,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	token := client.Publish(topic, cfg.QoS, false, payload)
	if token.WaitTimeout(10 * time.Second) && token.Error() != nil {
		return fmt.Errorf("failed to publish message: %w", token.Error())
	}

	return nil
}

// PublishBatch publishes multiple events to MQTT
func (p *Publisher) PublishBatch(evts []events.NormalizedEvent) error {
	for _, event := range evts {
		if err := p.Publish(event); err != nil {
			return err
		}
	}
	return nil
}

// IsConnected returns whether the publisher is ready
func (p *Publisher) IsConnected() bool {
	return p.client.IsConnected()
}
