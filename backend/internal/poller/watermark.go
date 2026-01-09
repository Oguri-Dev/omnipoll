package poller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	WatermarkPathEnv     = "OMNIPOLL_WATERMARK_PATH"
	DefaultWatermarkPath = "./data/watermark.json"
)

// Watermark tracks the last processed position for incremental polling
type Watermark struct {
	LastFechaHora      time.Time `json:"lastFechaHora"`
	IDsAtLastFechaHora []string  `json:"idsAtLastFechaHora"`
}

// WatermarkManager handles watermark persistence
type WatermarkManager struct {
	mu        sync.RWMutex
	watermark Watermark
	path      string
}

// NewWatermarkManager creates a new watermark manager
func NewWatermarkManager() *WatermarkManager {
	path := os.Getenv(WatermarkPathEnv)
	if path == "" {
		path = DefaultWatermarkPath
	}

	return &WatermarkManager{
		path: path,
		watermark: Watermark{
			LastFechaHora:      time.Time{},
			IDsAtLastFechaHora: []string{},
		},
	}
}

// Load reads the watermark from disk
func (m *WatermarkManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			// No watermark file yet, start fresh
			return nil
		}
		return err
	}

	var wm Watermark
	if err := json.Unmarshal(data, &wm); err != nil {
		return err
	}

	m.watermark = wm
	return nil
}

// Save persists the watermark to disk
func (m *WatermarkManager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.watermark, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}

// Get returns the current watermark
func (m *WatermarkManager) Get() Watermark {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.watermark
}

// Update updates the watermark with new values
func (m *WatermarkManager) Update(fechaHora time.Time, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If new timestamp is greater, reset the IDs list
	if fechaHora.After(m.watermark.LastFechaHora) {
		m.watermark.LastFechaHora = fechaHora
		m.watermark.IDsAtLastFechaHora = ids
	} else if fechaHora.Equal(m.watermark.LastFechaHora) {
		// Same timestamp, append new IDs
		idSet := make(map[string]bool)
		for _, id := range m.watermark.IDsAtLastFechaHora {
			idSet[id] = true
		}
		for _, id := range ids {
			if !idSet[id] {
				m.watermark.IDsAtLastFechaHora = append(m.watermark.IDsAtLastFechaHora, id)
			}
		}
	}

	// Persist immediately
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.watermark, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}

// Reset clears the watermark
func (m *WatermarkManager) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watermark = Watermark{
		LastFechaHora:      time.Time{},
		IDsAtLastFechaHora: []string{},
	}

	// Delete the file if it exists
	if err := os.Remove(m.path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// GetPath returns the watermark file path
func (m *WatermarkManager) GetPath() string {
	return m.path
}
