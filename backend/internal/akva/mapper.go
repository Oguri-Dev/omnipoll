package akva

import (
	"time"

	"github.com/omnipoll/backend/internal/events"
)

// Mapper transforms SQL rows to normalized events

// ToNormalizedEvent converts a DetalleAlimentacion to a NormalizedEvent
func ToNormalizedEvent(record DetalleAlimentacion) events.NormalizedEvent {
	var diaStr string
	if !record.Dia.IsZero() {
		diaStr = record.Dia.Format("2006-01-02")
	}

	return events.NormalizedEvent{
		ID:            record.ID,
		Source:        "akva",
		Name:          record.Name,
		UnitName:      record.UnitName,
		FechaHora:     record.FechaHora.UTC().Format(time.RFC3339),
		Dia:           diaStr,
		Inicio:        record.Inicio,
		Fin:           record.Fin,
		Dif:           record.Dif,
		AmountGrams:   record.AmountGrams,
		PelletFishMin: record.PelletFishMin,
		FishCount:     record.FishCount,
		PesoProm:      record.PesoProm,
		Biomasa:       record.Biomasa,
		PelletPK:      record.PelletPK,
		FeedName:      record.FeedName,
		SiloName:      record.SiloName,
		DoserName:     record.DoserName,
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
