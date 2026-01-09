package mqtt

import (
	"encoding/json"
	"fmt"
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

// Publish publishes a single event to MQTT
func (p *Publisher) Publish(event events.NormalizedEvent) error {
	if !p.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	cfg := p.client.GetConfig()
	client := p.client.GetClient()

	token := client.Publish(cfg.Topic, cfg.QoS, false, payload)
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
