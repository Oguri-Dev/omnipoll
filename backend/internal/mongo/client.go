package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/omnipoll/backend/internal/config"
)

// Client manages MongoDB connection
type Client struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	config     config.MongoDBConfig
}

// NewClient creates a new MongoDB client
func NewClient(cfg config.MongoDBConfig) *Client {
	return &Client{
		config: cfg,
	}
}

// Connect establishes connection to MongoDB
func (c *Client) Connect(ctx context.Context) error {
	clientOpts := options.Client().
		ApplyURI(c.config.URI).
		SetConnectTimeout(5 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.client = client
	c.database = client.Database(c.config.Database)
	c.collection = c.database.Collection(c.config.Collection)

	// Create indexes
	if err := c.createIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes creates the recommended indexes
func (c *Client) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: map[string]int{"fechaHora": 1},
		},
		{
			Keys: map[string]int{"unitName": 1},
		},
		{
			Keys: map[string]int{"source": 1},
		},
		{
			Keys: map[string]int{"ingestedAt": 1},
		},
	}

	_, err := c.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Disconnect closes the MongoDB connection
func (c *Client) Disconnect(ctx context.Context) error {
	if c.client != nil {
		return c.client.Disconnect(ctx)
	}
	return nil
}

// IsConnected checks if connection is alive
func (c *Client) IsConnected() bool {
	if c.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.Ping(ctx, readpref.Primary()) == nil
}

// GetCollection returns the events collection
func (c *Client) GetCollection() *mongo.Collection {
	return c.collection
}

// TestConnection tests the MongoDB connection
func (c *Client) TestConnection(ctx context.Context) error {
	if c.client == nil {
		if err := c.Connect(ctx); err != nil {
			return err
		}
	}
	return c.client.Ping(ctx, readpref.Primary())
}
