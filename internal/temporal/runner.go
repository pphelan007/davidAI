package temporal

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/pphelan007/davidAI/internal/config"
	"github.com/pphelan007/davidAI/internal/database"
	"github.com/pphelan007/davidAI/internal/temporal/activities"
)

// Runner wraps the activity client and manages the Temporal worker lifecycle
type Runner struct {
	client           client.Client
	activitiesClient *activities.ActivitiesClient
	worker           *Worker
	cfg              *config.TemporalConfig
}

// NewRunner creates a new runner instance with Temporal client, activity client, and worker
func NewRunner(ctx context.Context, cfg *config.TemporalConfig, dbCfg *config.DatabaseConfig) (*Runner, error) {
	// Create database client
	dbClient, err := database.NewClient(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %w", err)
	}

	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.Address,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		dbClient.Close()
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	// Create activities client
	activitiesClient := activities.NewActivitiesClient(ctx, temporalClient, dbClient)

	// Create worker
	worker, err := NewWorker(temporalClient, cfg.TaskQueue, dbClient)
	if err != nil {
		temporalClient.Close()
		dbClient.Close()
		return nil, fmt.Errorf("failed to create worker: %w", err)
	}

	return &Runner{
		client:           temporalClient,
		activitiesClient: activitiesClient,
		worker:           worker,
		cfg:              cfg,
	}, nil
}

// Start starts the runner and worker
func (r *Runner) Start() error {
	log.Printf("Starting Temporal runner on task queue: %s", r.cfg.TaskQueue)
	r.worker.Start(r.cfg.WorkerCount)
	return nil
}

// Stop stops the runner and worker, and closes the client
func (r *Runner) Stop() error {
	log.Println("Stopping Temporal runner...")

	if r.worker != nil {
		if err := r.worker.Stop(); err != nil {
			log.Printf("Error stopping worker: %v", err)
		}
	}

	if r.client != nil {
		r.client.Close()
	}

	log.Println("Temporal runner stopped")
	return nil
}

// GetActivitiesClient returns the activities client
func (r *Runner) GetActivitiesClient() *activities.ActivitiesClient {
	return r.activitiesClient
}

// GetClient returns the Temporal client
func (r *Runner) GetClient() client.Client {
	return r.client
}

// GetWorker returns the worker instance
func (r *Runner) GetWorker() *Worker {
	return r.worker
}
