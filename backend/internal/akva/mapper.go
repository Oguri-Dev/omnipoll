package akva

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/omnipoll/backend/internal/events"
)

// sanitizeString removes invalid UTF-8 characters
func sanitizeString(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// Remove invalid UTF-8 sequences
	return strings.ToValidUTF8(s, "")
}

// Mapper transforms SQL rows to normalized events

// ToNormalizedEvent converts a DetalleAlimentacion to a NormalizedEvent
func ToNormalizedEvent(record DetalleAlimentacion) events.NormalizedEvent {
	var diaStr string
	if !record.Dia.IsZero() {
		diaStr = record.Dia.Format("2006-01-02")
	}

	return events.NormalizedEvent{
		ID:            sanitizeString(record.ID),
		Source:        "akva",
		Name:          sanitizeString(record.Name),
		UnitName:      sanitizeString(record.UnitName),
		FechaHora:     record.FechaHora.UTC().Format(time.RFC3339),
		Dia:           diaStr,
		Inicio:        sanitizeString(record.Inicio),
		Fin:           sanitizeString(record.Fin),
		Dif:           record.Dif,
		AmountGrams:   record.AmountGrams,
		PelletFishMin: record.PelletFishMin,
		FishCount:     record.FishCount,
		PesoProm:      record.PesoProm,
		Biomasa:       record.Biomasa,
		PelletPK:      record.PelletPK,
		FeedName:      sanitizeString(record.FeedName),
		SiloName:      sanitizeString(record.SiloName),
		DoserName:     sanitizeString(record.DoserName),
		GramsPerSec:   record.GramsPerSec,
		KgTonMin:      record.KgTonMin,
		Marca:         record.Marca,
		IngestedAt:    time.Now().UTC(),
	}
}

// ToNormalizedEvents converts a slice of records to events
func ToNormalizedEvents(records []DetalleAlimentacion) []events.NormalizedEvent {
	evts := make([]events.NormalizedEvent, len(records))
	for i, r := range records {
		evts[i] = ToNormalizedEvent(r)
	}
	return evts
}
