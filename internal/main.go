// Package internal provides the main worker application logic.
package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/pphelan007/davidAI/internal/config"
	"github.com/pphelan007/davidAI/internal/database"
	"github.com/pphelan007/davidAI/internal/temporal"
	"github.com/pphelan007/davidAI/internal/temporal/activities"
	"github.com/pphelan007/davidAI/internal/utils"
)

// Run starts the worker and blocks until shutdown
func Run(cfg *config.Config) error {
	// 1. Setup Logging
	// Using standard log package for now, can be upgraded to zerolog later
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 2. Create Database Client
	dbClient, err := database.NewClient(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create database client: %w", err)
	}
	defer dbClient.Close()

	// 3. Create Temporal Client
	temporalClient, err := temporal.NewTemporalClient(context.Background(), cfg.Temporal.Address, cfg.Temporal.Namespace)
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}
	defer temporalClient.Close()

	// 4. Create the Worker Object
	worker, err := temporal.NewWorker(temporalClient.GetClient(), cfg.Temporal.TaskQueue, dbClient)
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	// 5. Create Activities Client
	activitiesClient := activities.NewActivitiesClient(context.Background(), temporalClient.GetClient(), dbClient)

	// 6. Start Worker Routine (closure captures activitiesClient)
	workerRoutine := utils.NewWorkerRoutine(
		"worker",
		func() {
			worker.Start(activitiesClient)
		},
		worker.Stop,
	)

	mainWg, closeables, startErr := utils.StartRoutines([]utils.Routine{
		workerRoutine,
	})

	if startErr != nil {
		return fmt.Errorf("failed to start routines: %w", startErr)
	}

	// 7. Log That Worker Started
	log.Println("Worker started")
	log.Println("Worker running, waiting for shutdown signal...")

	// 8. Block Until Shutdown
	mainWg.Wait()

	// 9. Cleanup (Reverse Order)
	for i := len(closeables) - 1; i >= 0; i-- {
		if err := closeables[i].Close(); err != nil {
			log.Printf("Error closing %T: %v", closeables[i], err)
		}
	}

	// 10. Log That Worker Stopped
	log.Println("Worker stopped")

	return nil
}
