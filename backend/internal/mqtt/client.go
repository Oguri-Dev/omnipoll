package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
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
	stopHeartbeat chan struct{}
}

// NewClient creates a new MQTT client
func NewClient(cfg config.MQTTConfig) *Client {
	return &Client{
		config: cfg,
		stopHeartbeat: make(chan struct{}),
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
	
	// Recreate stopHeartbeat channel for this connection
	c.stopHeartbeat = make(chan struct{})
	
	// Start heartbeat goroutine
	go c.startHeartbeat()
	fmt.Printf("[MQTT Client] Heartbeat started\n")
	
	return nil
}

// Disconnect closes the MQTT connection
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop heartbeat goroutine if channel is open
	select {
	case <-c.stopHeartbeat:
		// Already closed
	default:
		close(c.stopHeartbeat)
	}

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

// startHeartbeat sends periodic heartbeat messages to MQTT broker
func (c *Client) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.sendHeartbeat()
		case <-c.stopHeartbeat:
			log.Printf("[MQTT Heartbeat] Stopped")
			return
		}
	}
}

// sendHeartbeat sends a single heartbeat message
func (c *Client) sendHeartbeat() {
	if c.client == nil || !c.client.IsConnected() {
		return
	}

	heartbeat := map[string]interface{}{
		"status":    "connected",
		"timestamp": time.Now().Format(time.RFC3339),
		"clientId":  c.config.ClientID,
		"version":   "1.0",
	}

	payload, err := json.Marshal(heartbeat)
	if err != nil {
		return
	}

	topic := "feeding/mowi/status"
	
	// Fire-and-forget with QoS 0
	token := c.client.Publish(topic, 0, false, payload)
	token.Wait() // Don't block - just ensure message is queued
	
	log.Printf("[MQTT Heartbeat] ✓ Sent to %s", topic)
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
