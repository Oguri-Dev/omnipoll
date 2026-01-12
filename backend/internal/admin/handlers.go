package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/omnipoll/backend/internal/config"
	"github.com/omnipoll/backend/internal/events"
)

// StatusResponse represents the system status
type StatusResponse struct {
	WorkerRunning  bool              `json:"workerRunning"`
	LastFechaHora  string            `json:"lastFechaHora"`
	EventsToday    int64             `json:"eventsToday"`
	IngestionRate  float64           `json:"ingestionRate"`
	TotalEvents    int64             `json:"totalEvents"`
	Connections    ConnectionsStatus `json:"connections"`
	UptimeSeconds  int64             `json:"uptimeSeconds"`
}

type ConnectionsStatus struct {
	SQLServer bool `json:"sqlServer"`
	MQTT      bool `json:"mqtt"`
	MongoDB   bool `json:"mongodb"`
}

var startTime = time.Now()

// handleStatus returns system status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var lastFechaHora string
	var eventsToday, totalEvents int64
	var ingestionRate float64
	var sqlConnected, mqttConnected, mongoConnected bool
	var workerRunning bool

	if s.worker != nil {
		workerRunning = s.worker.IsRunning()
		stats := s.worker.GetStats()
		if !stats.LastFechaHora.IsZero() {
			lastFechaHora = stats.LastFechaHora.Format(time.RFC3339)
		}
		eventsToday = stats.EventsToday
		totalEvents = stats.TotalEvents
		ingestionRate = stats.IngestionRate
		sqlConnected = stats.SQLConnected
		mqttConnected = stats.MQTTConnected
		mongoConnected = stats.MongoConnected
	}

	resp := StatusResponse{
		WorkerRunning:  workerRunning,
		LastFechaHora:  lastFechaHora,
		EventsToday:    eventsToday,
		IngestionRate:  ingestionRate,
		TotalEvents:    totalEvents,
		Connections: ConnectionsStatus{
			SQLServer: sqlConnected,
			MQTT:      mqttConnected,
			MongoDB:   mongoConnected,
		},
		UptimeSeconds: int64(time.Since(startTime).Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding status response: %v", err)
	}
}

// handleConfig handles GET and PUT for configuration
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.configManager.Get()
		// Mask passwords in response
		cfg.SQLServer.Password = maskPassword(cfg.SQLServer.Password)
		cfg.MQTT.Password = maskPassword(cfg.MQTT.Password)
		cfg.Admin.Password = maskPassword(cfg.Admin.Password)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)

	case http.MethodPut:
		var cfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		// If password is masked, keep the original
		currentCfg := s.configManager.Get()
		if cfg.SQLServer.Password == "********" {
			cfg.SQLServer.Password = currentCfg.SQLServer.Password
		}
		if cfg.MQTT.Password == "********" {
			cfg.MQTT.Password = currentCfg.MQTT.Password
		}
		if cfg.Admin.Password == "********" {
			cfg.Admin.Password = currentCfg.Admin.Password
		}

		if err := s.configManager.Update(cfg); err != nil {
			http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWorkerStart starts the polling worker
func (s *Server) handleWorkerStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.worker == nil {
		http.Error(w, "Worker not initialized", http.StatusInternalServerError)
		return
	}

	if s.worker.IsRunning() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "already_running"})
		return
	}

	if err := s.worker.Start(); err != nil {
		http.Error(w, "Failed to start worker: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// handleWorkerStop stops the polling worker
func (s *Server) handleWorkerStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.worker == nil {
		http.Error(w, "Worker not initialized", http.StatusInternalServerError)
		return
	}

	if !s.worker.IsRunning() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "already_stopped"})
		return
	}

	s.worker.Stop()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// handleWatermarkReset resets the watermark to start fresh
func (s *Server) handleWatermarkReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.worker == nil {
		http.Error(w, "Worker not initialized", http.StatusInternalServerError)
		return
	}

	if s.worker.IsRunning() {
		http.Error(w, "Stop worker before resetting watermark", http.StatusBadRequest)
		return
	}

	if err := s.worker.ResetWatermark(); err != nil {
		http.Error(w, "Failed to reset watermark: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
}

// handleTestSQLServer tests SQL Server connection
func (s *Server) handleTestSQLServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.worker == nil {
		http.Error(w, "Worker not initialized", http.StatusInternalServerError)
		return
	}

	ok, err := s.worker.TestSQLConnection()
	result := map[string]interface{}{"connected": ok}
	if err != nil {
		result["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleTestMQTT tests MQTT connection
func (s *Server) handleTestMQTT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.worker == nil {
		http.Error(w, "Worker not initialized", http.StatusInternalServerError)
		return
	}

	ok, err := s.worker.TestMQTTConnection()
	result := map[string]interface{}{"connected": ok}
	if err != nil {
		result["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleTestMongoDB tests MongoDB connection
func (s *Server) handleTestMongoDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.worker == nil {
		http.Error(w, "Worker not initialized", http.StatusInternalServerError)
		return
	}

	ok, err := s.worker.TestMongoConnection()
	result := map[string]interface{}{"connected": ok}
	if err != nil {
		result["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleLogs returns recent logs
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, return worker logs if available
	var logs []events.LogEntry
	if s.worker != nil {
		logs = s.worker.GetLogs()
	}

	if logs == nil {
		logs = []events.LogEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func maskPassword(password string) string {
	if password == "" {
		return ""
	}
	return "********"
}

// handleEvents returns recent events from MongoDB
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.worker == nil {
		http.Error(w, "Worker not initialized", http.StatusInternalServerError)
		return
	}

	events, err := s.worker.GetRecentEvents(r.Context(), 100)
	if err != nil {
		http.Error(w, "Failed to get events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Transform MongoDB HistoricalEvent to frontend-friendly format
	response := make([]map[string]interface{}, 0, len(events))
	for _, evt := range events {
		response = append(response, map[string]interface{}{
			"_id":        evt.ID,
			"source":     evt.Source,
			"fechaHora":  evt.FechaHora.Format(time.RFC3339),
			"unitName":   evt.UnitName,
			"payload":    evt.Payload,
			"ingestedAt": evt.IngestedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
