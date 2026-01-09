package events

import "time"

// NormalizedEvent represents the stable JSON event format
type NormalizedEvent struct {
	ID            string    `json:"id"`
	Source        string    `json:"source"`
	Name          string    `json:"name"`
	UnitName      string    `json:"unitName"`
	FechaHora     string    `json:"fechaHora"` // RFC3339 UTC
	Dia           string    `json:"dia"`       // Date only
	Inicio        string    `json:"inicio"`
	Fin           string    `json:"fin"`
	Dif           int       `json:"dif"`
	AmountGrams   float64   `json:"amountGrams"`
	PelletFishMin float64   `json:"pelletFishMin"`
	FishCount     float64   `json:"fishCount"`
	PesoProm      float64   `json:"pesoProm"`
	Biomasa       float64   `json:"biomasa"`
	PelletPK      float64   `json:"pelletPK"`
	FeedName      string    `json:"feedName"`
	SiloName      string    `json:"siloName"`
	DoserName     string    `json:"doserName"`
	GramsPerSec   float64   `json:"gramsPerSec"`
	KgTonMin      float64   `json:"kgTonMin"`
	Marca         int       `json:"marca"`
	IngestedAt    time.Time `json:"ingestedAt"`
}

// LogEntry represents a log entry (shared type to avoid import cycles)
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}
