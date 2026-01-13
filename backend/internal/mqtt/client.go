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

	// Determine protocol: use TLS for port 8883 or if UseTLS is explicitly set
	protocol := "tcp"
	if c.config.UseTLS || c.config.Port == 8883 {
		protocol = "ssl"
	}
	broker := fmt.Sprintf("%s://%s:%d", protocol, c.config.Broker, c.config.Port)

	fmt.Printf("[MQTT Client] Connecting to broker: %s (User: %s, TLS: %v)\n", broker, c.config.User, c.config.UseTLS)

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
			fmt.Printf("[MQTT Client] Connection lost: %v\n", err)
		}).
		SetOnConnectHandler(func(client paho.Client) {
			c.mu.Lock()
			c.connected = true
			c.mu.Unlock()
			fmt.Printf("[MQTT Client] Connected successfully to %s\n", broker)
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
		return fmt.Errorf("failed to connect to MQTT broker %s: %w", broker, token.Error())
	}

	c.client = client
	c.connected = true
	fmt.Printf("[MQTT Client] Connection established to %s\n", broker)
	
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
