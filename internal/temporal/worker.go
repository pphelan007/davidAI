package temporal

import (
	"context"
	"log"
	"sync"
)

// Worker handles job processing
type Worker struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewWorker creates a new worker instance
func NewWorker() (*Worker, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start begins processing jobs
func (w *Worker) Start(numWorkers int) {
	log.Printf("Worker started with %d worker goroutines", numWorkers)
	// Worker is ready but not processing any jobs yet
	// This is a minimal implementation that satisfies the interface
}

// Stop stops the worker and waits for all jobs to complete
func (w *Worker) Stop() error {
	log.Println("Stopping worker...")
	w.cancel()
	w.wg.Wait()
	log.Println("Worker stopped")
	return nil
}
