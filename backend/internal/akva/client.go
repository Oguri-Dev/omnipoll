package akva

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/omnipoll/backend/internal/config"
)

// Client handles SQL Server connection to Akva database
type Client struct {
	db     *sql.DB
	config config.SQLServerConfig
}

// NewClient creates a new Akva SQL Server client
func NewClient(cfg config.SQLServerConfig) *Client {
	return &Client{
		config: cfg,
	}
}

// Connect establishes connection to SQL Server
func (c *Client) Connect(ctx context.Context) error {
	connString := fmt.Sprintf(
		"server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=disable;connection timeout=3",
		c.config.Host,
		c.config.Port,
		c.config.Database,
		c.config.User,
		c.config.Password,
	)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.db = db
	return nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// IsConnected checks if connection is alive
func (c *Client) IsConnected() bool {
	if c.db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.db.PingContext(ctx) == nil
}

// FetchNewRecords fetches records newer than the watermark
func (c *Client) FetchNewRecords(ctx context.Context, lastFechaHora time.Time, seenIDs []string, batchSize int) ([]DetalleAlimentacion, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// If watermark is zero/empty (fresh start), use a date far in the past to get oldest records
	// Otherwise use the watermark timestamp to find new records
	queryTimestamp := lastFechaHora
	if lastFechaHora.IsZero() || (lastFechaHora.Year() == 1 && lastFechaHora.Month() == 1) {
		// Fresh start - query from year 2000 onwards
		queryTimestamp = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	query := `
		SELECT TOP (@batchSize)
			ID,
			Name,
			UnitName,
			FechaHora,
			Dia,
			inicio,
			Fin,
			dif,
			AmountGrams,
			pelletfishmin,
			FisCount,
			PesoProm,
			Biomasa,
			pelletpK,
			Feedname,
			SiloName,
			DoserName,
			gramspersec,
			kgtonmin,
			Marca
		FROM dbo.TB_DetalleAlimentacion
		WHERE FechaHora >= @lastFechaHora
		ORDER BY FechaHora ASC, ID ASC
	`

	rows, err := c.db.QueryContext(ctx, query,
		sql.Named("batchSize", batchSize),
		sql.Named("lastFechaHora", queryTimestamp),
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Build a set of seen IDs for quick lookup
	seenSet := make(map[string]bool)
	for _, id := range seenIDs {
		seenSet[id] = true
	}

	var records []DetalleAlimentacion
	for rows.Next() {
		var r DetalleAlimentacion
		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.UnitName,
			&r.FechaHora,
			&r.Dia,
			&r.Inicio,
			&r.Fin,
			&r.Dif,
			&r.AmountGrams,
			&r.PelletFishMin,
			&r.FishCount,
			&r.PesoProm,
			&r.Biomasa,
			&r.PelletPK,
			&r.FeedName,
			&r.SiloName,
			&r.DoserName,
			&r.GramsPerSec,
			&r.KgTonMin,
			&r.Marca,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		// Skip already-seen records at the same timestamp
		if r.FechaHora.Equal(lastFechaHora) && seenSet[r.ID] {
			continue
		}

		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return records, nil
}

// TestConnection tests the SQL Server connection
func (c *Client) TestConnection(ctx context.Context) error {
	if c.db == nil {
		// Try to connect
		if err := c.Connect(ctx); err != nil {
			return err
		}
	}
	return c.db.PingContext(ctx)
}
