// Package temporal provides Temporal workflow and activity management.
package temporal

import (
	"context"
	"log"
	"sync"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/pphelan007/davidAI/internal/database"
	"github.com/pphelan007/davidAI/internal/temporal/activities"
	"github.com/pphelan007/davidAI/internal/temporal/workflows"
)

// Worker handles Temporal job processing
type Worker struct {
	client         client.Client
	temporalWorker worker.Worker
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	taskQueue      string
	dbClient       *database.Client
}

// NewWorker creates a new worker instance
func NewWorker(c client.Client, taskQueue string, dbClient *database.Client) (*Worker, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create Temporal worker
	temporalWorker := worker.New(c, taskQueue, worker.Options{})

	return &Worker{
		client:         c,
		temporalWorker: temporalWorker,
		ctx:            ctx,
		cancel:         cancel,
		taskQueue:      taskQueue,
		dbClient:       dbClient,
	}, nil
}

// RegisterWorkflows registers all workflows with the worker
func (w *Worker) RegisterWorkflows() {
	workflows.RegisterWorkflows(w.temporalWorker)
	log.Println("Workflows registered")
}

// RegisterActivities registers all activities with the worker
func (w *Worker) RegisterActivities(activitiesClient *activities.ActivitiesClient) {
	activities.RegisterActivities(w.temporalWorker, activitiesClient)
	log.Println("Activities registered")
}

// Start begins processing jobs
func (w *Worker) Start(numWorkers int) {
	log.Printf("Starting Temporal worker with %d worker goroutines on task queue: %s", numWorkers, w.taskQueue)

	// Register workflows and activities
	w.RegisterWorkflows()

	// Create activities client and register activities
	activitiesClient := activities.NewActivitiesClient(w.ctx, w.client, w.dbClient)
	w.RegisterActivities(activitiesClient)

	// Start the worker in a goroutine
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		// Create interrupt channel that closes when context is cancelled
		interruptCh := make(chan interface{})
		go func() {
			<-w.ctx.Done()
			close(interruptCh)
		}()
		if err := w.temporalWorker.Run(interruptCh); err != nil {
			log.Printf("Temporal worker error: %v", err)
		}
	}()

	log.Println("Temporal worker started and ready to process tasks")
}

// Stop stops the worker and waits for all jobs to complete
func (w *Worker) Stop() error {
	log.Println("Stopping Temporal worker...")
	w.cancel()
	w.temporalWorker.Stop()
	w.wg.Wait()
	if w.dbClient != nil {
		w.dbClient.Close()
	}
	log.Println("Temporal worker stopped")
	return nil
}
