package mqtt

import (
	"fmt"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/omnipoll/backend/internal/config"
)

// Client manages MQTT broker connection
type Client struct {
	mu       sync.RWMutex
	client   paho.Client
	config   config.MQTTConfig
	connected bool
}

// NewClient creates a new MQTT client
func NewClient(cfg config.MQTTConfig) *Client {
	return &Client{
		config: cfg,
	}
}

// Connect establishes connection to MQTT broker
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	broker := fmt.Sprintf("tcp://%s:%d", c.config.Broker, c.config.Port)

	opts := paho.NewClientOptions().
		AddBroker(broker).
		SetClientID(c.config.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetCleanSession(true).
		SetConnectionLostHandler(func(client paho.Client, err error) {
			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()
		}).
		SetOnConnectHandler(func(client paho.Client) {
			c.mu.Lock()
			c.connected = true
			c.mu.Unlock()
		})

	if c.config.User != "" {
		opts.SetUsername(c.config.User)
	}
	if c.config.Password != "" {
		opts.SetPassword(c.config.Password)
	}

	client := paho.NewClient(opts)

	token := client.Connect()
	if token.WaitTimeout(5 * time.Second) && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	c.client = client
	c.connected = true
	return nil
}

// Disconnect closes the MQTT connection
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(1000)
	}
	c.connected = false
}

// IsConnected returns the connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		return false
	}
	return c.client.IsConnected()
}

// GetClient returns the underlying paho client
func (c *Client) GetClient() paho.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

// GetConfig returns the MQTT configuration
func (c *Client) GetConfig() config.MQTTConfig {
	return c.config
}

// TestConnection tests the MQTT connection
func (c *Client) TestConnection() error {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	if !c.client.IsConnected() {
		return fmt.Errorf("not connected to MQTT broker")
	}

	return nil
}
