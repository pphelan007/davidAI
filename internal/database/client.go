// Package database provides database client and operations
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/pphelan007/davidAI/internal/config"
)

// Client wraps the database connection
type Client struct {
	DB *sql.DB
}

// Asset represents an asset record in the database
type Asset struct {
	ID            string
	WorkflowID    string
	WorkflowRunID string
	ParentAssetID *string
	FilePath      string
	ContentHash   string
	AssetType     string
	CreatedAt     time.Time
}

// NewClient creates a new database client
func NewClient(cfg *config.DatabaseConfig) (*Client, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	client := &Client{DB: db}

	// Initialize schema
	if err := client.InitSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return client, nil
}

// InitSchema creates the database tables if they don't exist
func (c *Client) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS assets (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		workflow_id VARCHAR(255) NOT NULL,
		workflow_run_id VARCHAR(255) NOT NULL,
		parent_asset_id UUID REFERENCES assets(id),
		file_path TEXT NOT NULL,
		content_hash VARCHAR(64) NOT NULL,
		asset_type VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_workflow_run ON assets(workflow_run_id);
	CREATE INDEX IF NOT EXISTS idx_parent_asset ON assets(parent_asset_id);
	CREATE INDEX IF NOT EXISTS idx_content_hash ON assets(content_hash);
	CREATE INDEX IF NOT EXISTS idx_workflow_id ON assets(workflow_id);
	`

	_, err := c.DB.Exec(query)
	return err
}

// InsertAsset inserts a new asset record into the database
func (c *Client) InsertAsset(asset *Asset) error {
	query := `
	INSERT INTO assets (id, workflow_id, workflow_run_id, parent_asset_id, file_path, content_hash, asset_type, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := c.DB.Exec(
		query,
		asset.ID,
		asset.WorkflowID,
		asset.WorkflowRunID,
		asset.ParentAssetID,
		asset.FilePath,
		asset.ContentHash,
		asset.AssetType,
		asset.CreatedAt,
	)

	return err
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.DB.Close()
}
