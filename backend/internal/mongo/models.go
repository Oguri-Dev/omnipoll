package mongo

import "time"

// HistoricalEvent represents a document in MongoDB
type HistoricalEvent struct {
	ID         string                 `bson:"_id"`
	Source     string                 `bson:"source"`
	FechaHora  time.Time              `bson:"fechaHora"`
	UnitName   string                 `bson:"unitName"`
	Payload    map[string]interface{} `bson:"payload"`
	IngestedAt time.Time              `bson:"ingestedAt"`
}
