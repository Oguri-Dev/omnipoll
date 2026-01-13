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
