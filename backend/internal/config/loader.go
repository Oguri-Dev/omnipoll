package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/omnipoll/backend/internal/crypto"
	"gopkg.in/yaml.v3"
)

const (
	ConfigPathEnv     = "OMNIPOLL_CONFIG_PATH"
	DefaultConfigPath = "./data/config.yaml"
)

// Manager handles configuration loading, saving, and encryption
type Manager struct {
	mu        sync.RWMutex
	config    *Config
	path      string
	encryptor *crypto.Encryptor
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	encryptor, err := crypto.NewEncryptor()
	if err != nil {
		return nil, err
	}

	path := os.Getenv(ConfigPathEnv)
	if path == "" {
		path = DefaultConfigPath
	}

	m := &Manager{
		path:      path,
		encryptor: encryptor,
		config:    DefaultConfig(),
	}

	return m, nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		SQLServer: SQLServerConfig{
			Host:     "localhost",
			Port:     1433,
			Database: "FTFeeding",
			User:     "sa",
			Password: "",
		},
		MQTT: MQTTConfig{
			Broker:   "localhost",
			Port:     1883,
			Topic:    "ftfeeding/akva/detalle",
			ClientID: "omnipoll-worker",
			QoS:      1,
		},
		MongoDB: MongoDBConfig{
			URI:        "mongodb://localhost:27017",
			Database:   "omnipoll",
			Collection: "historical_events",
		},
		Polling: PollingConfig{
			IntervalMS: 5000,
			BatchSize:  100,
		},
		Admin: AdminConfig{
			Host:     "127.0.0.1",
			Port:     8080,
			Username: "admin",
			Password: "admin",
		},
	}
}

// Load reads the configuration from file
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if it doesn't exist
			return m.saveUnlocked()
		}
		return err
	}

	ext := filepath.Ext(m.path)
	var cfg Config

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return err
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return err
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			if err := json.Unmarshal(data, &cfg); err != nil {
				return err
			}
		}
	}

	// Decrypt passwords
	if cfg.SQLServer.Password, err = m.encryptor.Decrypt(cfg.SQLServer.Password); err != nil {
		return err
	}
	if cfg.MQTT.Password, err = m.encryptor.Decrypt(cfg.MQTT.Password); err != nil {
		return err
	}
	if cfg.Admin.Password, err = m.encryptor.Decrypt(cfg.Admin.Password); err != nil {
		return err
	}

	m.config = &cfg
	return nil
}

// Save writes the configuration to file with encrypted passwords
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveUnlocked()
}

func (m *Manager) saveUnlocked() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create a copy with encrypted passwords
	cfg := *m.config

	var err error
	if !crypto.IsEncrypted(cfg.SQLServer.Password) && cfg.SQLServer.Password != "" {
		cfg.SQLServer.Password, err = m.encryptor.Encrypt(cfg.SQLServer.Password)
		if err != nil {
			return err
		}
	}
	if !crypto.IsEncrypted(cfg.MQTT.Password) && cfg.MQTT.Password != "" {
		cfg.MQTT.Password, err = m.encryptor.Encrypt(cfg.MQTT.Password)
		if err != nil {
			return err
		}
	}
	if !crypto.IsEncrypted(cfg.Admin.Password) && cfg.Admin.Password != "" {
		cfg.Admin.Password, err = m.encryptor.Encrypt(cfg.Admin.Password)
		if err != nil {
			return err
		}
	}

	ext := filepath.Ext(m.path)
	var data []byte

	switch ext {
	case ".json":
		data, err = json.MarshalIndent(cfg, "", "  ")
	default:
		data, err = yaml.Marshal(cfg)
	}
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0600)
}

// Get returns the current configuration (read-only copy)
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.config
}

// Update updates the configuration
func (m *Manager) Update(cfg Config) error {
	m.mu.Lock()
	m.config = &cfg
	m.mu.Unlock()
	return m.Save()
}

// GetPath returns the configuration file path
func (m *Manager) GetPath() string {
	return m.path
}
