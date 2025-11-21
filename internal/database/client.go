// Package database provides database client and operations
package database

import (
	"database/sql"
	"encoding/json"
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
	CreatedAt     time.Time
}

// Feature represents a feature record in the database
type Feature struct {
	ID                string
	AssetID           string
	FeatureType       string
	FeatureData       map[string]interface{} // Will be stored as JSONB
	ComputationParams map[string]interface{} // Will be stored as JSONB (optional)
	ComputedAt        time.Time
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
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_workflow_run ON assets(workflow_run_id);
	CREATE INDEX IF NOT EXISTS idx_parent_asset ON assets(parent_asset_id);
	CREATE INDEX IF NOT EXISTS idx_content_hash ON assets(content_hash);
	CREATE INDEX IF NOT EXISTS idx_workflow_id ON assets(workflow_id);

	CREATE TABLE IF NOT EXISTS features (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
		feature_type VARCHAR(50) NOT NULL,
		feature_data JSONB NOT NULL,
		computation_params JSONB,
		computed_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_features_asset ON features(asset_id);
	CREATE INDEX IF NOT EXISTS idx_features_type ON features(feature_type);
	CREATE INDEX IF NOT EXISTS idx_features_computed_at ON features(computed_at);
	`

	_, err := c.DB.Exec(query)
	return err
}

// InsertAsset inserts a new asset record into the database
func (c *Client) InsertAsset(asset *Asset) error {
	query := `
	INSERT INTO assets (id, workflow_id, workflow_run_id, parent_asset_id, file_path, content_hash, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := c.DB.Exec(
		query,
		asset.ID,
		asset.WorkflowID,
		asset.WorkflowRunID,
		asset.ParentAssetID,
		asset.FilePath,
		asset.ContentHash,
		asset.CreatedAt,
	)

	return err
}

// InsertFeature inserts a new feature record into the database
func (c *Client) InsertFeature(feature *Feature) error {
	// Convert feature data to JSON
	featureDataJSON, err := json.Marshal(feature.FeatureData)
	if err != nil {
		return fmt.Errorf("failed to marshal feature data: %w", err)
	}

	// Convert computation params to JSON (may be nil)
	var computationParamsJSON []byte
	if len(feature.ComputationParams) > 0 {
		computationParamsJSON, err = json.Marshal(feature.ComputationParams)
		if err != nil {
			return fmt.Errorf("failed to marshal computation params: %w", err)
		}
	}

	query := `
	INSERT INTO features (id, asset_id, feature_type, feature_data, computation_params, computed_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	var computationParams interface{}
	if len(computationParamsJSON) > 0 {
		computationParams = string(computationParamsJSON)
	}

	_, err = c.DB.Exec(
		query,
		feature.ID,
		feature.AssetID,
		feature.FeatureType,
		string(featureDataJSON),
		computationParams,
		feature.ComputedAt,
	)

	return err
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.DB.Close()
}
