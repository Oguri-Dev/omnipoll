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

// GetEventsByIDs returns events by their IDs for change comparison
func (r *Repository) GetEventsByIDs(ctx context.Context, source string, ids []string) (map[string]HistoricalEvent, error) {
	if len(ids) == 0 {
		return make(map[string]HistoricalEvent), nil
	}

	// Build MongoDB IDs (source:id format)
	mongoIDs := make([]string, len(ids))
	for i, id := range ids {
		mongoIDs[i] = fmt.Sprintf("%s:%s", source, id)
	}

	filter := bson.M{"_id": bson.M{"$in": mongoIDs}}
	cursor, err := r.client.GetCollection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]HistoricalEvent)
	for cursor.Next(ctx) {
		var event HistoricalEvent
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}
		// Use the original ID (without source prefix) as key
		result[event.ID] = event
	}

	return result, nil
}

// GetByID returns a single event by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*HistoricalEvent, error) {
	var event HistoricalEvent
	err := r.client.GetCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// QueryOptions defines filtering and pagination options
type QueryOptions struct {
	Page      int                  `json:"page"`      // 1-based page number
	PageSize  int                  `json:"pageSize"`  // Items per page
	StartDate *time.Time           `json:"startDate"` // Filter by date range
	EndDate   *time.Time           `json:"endDate"`
	Source    string               `json:"source"`    // Filter by source (Akva, etc) - deprecated, use Centro
	UnitName  string               `json:"unitName"`  // Filter by unit name - deprecated, use Jaula
	Centro    string               `json:"centro"`    // Filter by centro name
	Jaula     string               `json:"jaula"`     // Filter by jaula number
	SortBy    string               `json:"sortBy"`    // "fechaHora" or "ingestedAt"
	SortOrder int                  `json:"sortOrder"` // 1 for ascending, -1 for descending
}

// QueryResult represents paginated query results
type QueryResult struct {
	Data       []HistoricalEvent `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}

// QueryEvents retrieves events with filtering and pagination
func (r *Repository) QueryEvents(ctx context.Context, opts QueryOptions) (*QueryResult, error) {
	// Build filter
	filter := bson.M{}

	if opts.StartDate != nil || opts.EndDate != nil {
		dateFilter := bson.M{}
		if opts.StartDate != nil {
			dateFilter["$gte"] = opts.StartDate
		}
		if opts.EndDate != nil {
			dateFilter["$lte"] = opts.EndDate
		}
		filter["fechaHora"] = dateFilter
	}

	if opts.Source != "" {
		filter["source"] = opts.Source
	}

	// Filter by Centro (payload.name)
	if opts.Centro != "" {
		filter["payload.name"] = bson.M{"$regex": opts.Centro, "$options": "i"} // Case-insensitive
	}

	if opts.UnitName != "" {
		filter["unitName"] = bson.M{"$regex": opts.UnitName, "$options": "i"} // Case-insensitive
	}

	// Filter by Jaula (unitName)
	if opts.Jaula != "" {
		filter["unitName"] = bson.M{"$regex": opts.Jaula, "$options": "i"} // Case-insensitive
	}

	// Count total documents
	total, err := r.client.GetCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// Set default pagination values
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 50
	}
	if opts.PageSize > 500 {
		opts.PageSize = 500 // Max page size
	}

	// Set default sort
	if opts.SortBy == "" {
		opts.SortBy = "ingestedAt"
	}
	if opts.SortOrder != 1 {
		opts.SortOrder = -1
	}

	// Build query options
	skip := int64((opts.Page - 1) * opts.PageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(opts.PageSize)).
		SetSort(bson.D{{Key: opts.SortBy, Value: opts.SortOrder}})

	cursor, err := r.client.GetCollection().Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []HistoricalEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	if events == nil {
		events = []HistoricalEvent{}
	}

	totalPages := (int(total) + opts.PageSize - 1) / opts.PageSize

	return &QueryResult{
		Data:       events,
		Total:      total,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateByID updates an event
func (r *Repository) UpdateByID(ctx context.Context, id string, update map[string]interface{}) error {
	// Don't allow updating _id or ingestedAt
	delete(update, "_id")
	delete(update, "ingestedAt")

	result := r.client.GetCollection().FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if result.Err() != nil {
		return fmt.Errorf("failed to update event: %w", result.Err())
	}

	return nil
}

// DeleteByID deletes an event
func (r *Repository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.client.GetCollection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

// DeleteByFilter deletes events matching a filter
func (r *Repository) DeleteByFilter(ctx context.Context, source string, startDate *time.Time) (int64, error) {
	filter := bson.M{}

	if source != "" {
		filter["source"] = source
	}

	if startDate != nil {
		filter["ingestedAt"] = bson.M{"$lt": startDate}
	}

	result, err := r.client.GetCollection().DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete events: %w", err)
	}

	return result.DeletedCount, nil
}
