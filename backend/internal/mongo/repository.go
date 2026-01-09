package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/omnipoll/backend/internal/events"
)

// Repository handles MongoDB operations for historical persistence
type Repository struct {
	client *Client
}

// NewRepository creates a new MongoDB repository
func NewRepository(client *Client) *Repository {
	return &Repository{
		client: client,
	}
}

// Insert inserts a single event into MongoDB
func (r *Repository) Insert(ctx context.Context, event events.NormalizedEvent) error {
	doc := r.eventToDocument(event)

	_, err := r.client.GetCollection().InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// InsertBatch inserts multiple events into MongoDB
func (r *Repository) InsertBatch(ctx context.Context, evts []events.NormalizedEvent) error {
	if len(evts) == 0 {
		return nil
	}

	docs := make([]interface{}, len(evts))
	for i, event := range evts {
		docs[i] = r.eventToDocument(event)
	}

	opts := options.InsertMany().SetOrdered(false) // Continue on duplicate key errors
	_, err := r.client.GetCollection().InsertMany(ctx, docs, opts)
	if err != nil {
		// Check if it's a bulk write exception with only duplicate key errors
		// Those are expected and can be ignored for idempotency
		return fmt.Errorf("failed to insert events: %w", err)
	}

	return nil
}

// eventToDocument converts an event to a MongoDB document
func (r *Repository) eventToDocument(event events.NormalizedEvent) HistoricalEvent {
	fechaHora, _ := time.Parse(time.RFC3339, event.FechaHora)

	return HistoricalEvent{
		ID:        fmt.Sprintf("%s:%s", event.Source, event.ID),
		Source:    event.Source,
		FechaHora: fechaHora,
		UnitName:  event.UnitName,
		Payload: map[string]interface{}{
			"name":          event.Name,
			"dia":           event.Dia,
			"inicio":        event.Inicio,
			"fin":           event.Fin,
			"dif":           event.Dif,
			"amountGrams":   event.AmountGrams,
			"pelletFishMin": event.PelletFishMin,
			"fishCount":     event.FishCount,
			"pesoProm":      event.PesoProm,
			"biomasa":       event.Biomasa,
			"pelletPK":      event.PelletPK,
			"feedName":      event.FeedName,
			"siloName":      event.SiloName,
			"doserName":     event.DoserName,
			"gramsPerSec":   event.GramsPerSec,
			"kgTonMin":      event.KgTonMin,
			"marca":         event.Marca,
		},
		IngestedAt: event.IngestedAt,
	}
}

// CountEvents returns the total number of events
func (r *Repository) CountEvents(ctx context.Context) (int64, error) {
	return r.client.GetCollection().CountDocuments(ctx, bson.M{})
}

// CountEventsToday returns the number of events ingested today
func (r *Repository) CountEventsToday(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	filter := bson.M{
		"ingestedAt": bson.M{
			"$gte": startOfDay,
		},
	}

	return r.client.GetCollection().CountDocuments(ctx, filter)
}

// IsConnected returns connection status
func (r *Repository) IsConnected() bool {
	return r.client.IsConnected()
}

// GetRecentEvents returns the most recent events
func (r *Repository) GetRecentEvents(ctx context.Context, limit int) ([]HistoricalEvent, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "ingestedAt", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.client.GetCollection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []HistoricalEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}
