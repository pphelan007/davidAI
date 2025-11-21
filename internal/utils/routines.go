// Package utils provides utilities for managing concurrent operations,
// signal handling, and timeout detection in the application.
package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ExecutionContext manages context and waitgroup for concurrent operations
type ExecutionContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewExecutionContext creates a new execution context with a cancellable context
func NewExecutionContext() *ExecutionContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &ExecutionContext{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context returns the context
func (ec *ExecutionContext) Context() context.Context {
	return ec.ctx
}

// Cancel cancels the context
func (ec *ExecutionContext) Cancel() {
	ec.cancel()
}

// StartRoutineWithError starts a goroutine that can return an error
// The error is sent to the provided error channel
func (ec *ExecutionContext) StartRoutineWithError(name string, fn func(context.Context) error, errChan chan<- error) {
	ec.wg.Add(1)
	go func() {
		defer ec.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				if errChan != nil {
					errChan <- fmt.Errorf("panic in routine %s: %v", name, r)
				}
			}
		}()
		if err := fn(ec.ctx); err != nil {
			if errChan != nil {
				errChan <- fmt.Errorf("error in routine %s: %w", name, err)
			}
		}
	}()
}

// Closeable represents a resource that can be closed/cleaned up
type Closeable interface {
	Close() error
}

// Routine represents a background routine that can be started and stopped
type Routine interface {
	// Name returns the name of the routine
	Name() string
	// Start starts the routine and blocks until it completes or context is cancelled
	Start(ctx context.Context) error
	// Close performs cleanup for the routine
	Close() error
}

// StartRoutines starts all provided routines and returns a waitgroup, closeables, and error
func StartRoutines(routines []Routine) (*sync.WaitGroup, []Closeable, error) {
	ec := NewExecutionContext()
	mainWg := &sync.WaitGroup{}
	closeables := make([]Closeable, 0, len(routines))
	errChan := make(chan error, len(routines))

	// Start each routine
	for _, routine := range routines {
		routine := routine // capture loop variable
		mainWg.Add(1)
		closeables = append(closeables, routine)

		ec.StartRoutineWithError(routine.Name(), func(ctx context.Context) error {
			defer mainWg.Done()
			return routine.Start(ctx)
		}, errChan)
	}

	// Monitor for errors and trigger shutdown
	go func() {
		for err := range errChan {
			if err != nil {
				log.Printf("Routine error: %v", err)
				ec.Cancel()
			}
		}
	}()

	// Start signal handler to cancel context on interrupt
	sigCtx, sigCancel := SetupInterruptHandler()
	go func() {
		<-sigCtx.Done()
		log.Println("Interrupt signal received, initiating shutdown...")
		ec.Cancel()
		sigCancel()
	}()

	return mainWg, closeables, nil
}

// WorkerRoutine implements Routine for a worker that needs to be started and stopped
type WorkerRoutine struct {
	name    string
	start   func(int) // Start function that takes number of workers (doesn't return error)
	stop    func() error
	workers int
}

// NewWorkerRoutine creates a new worker routine
func NewWorkerRoutine(name string, start func(int), stop func() error, numWorkers int) *WorkerRoutine {
	return &WorkerRoutine{
		name:    name,
		start:   start,
		stop:    stop,
		workers: numWorkers,
	}
}

// Name returns the routine name
func (r *WorkerRoutine) Name() string {
	return r.name
}

// Start starts the worker
func (r *WorkerRoutine) Start(ctx context.Context) error {
	log.Printf("Starting %s with %d worker goroutines", r.name, r.workers)
	r.start(r.workers)

	// Wait for context cancellation
	<-ctx.Done()
	log.Printf("Stopping %s...", r.name)
	return nil
}

// Close performs cleanup
func (r *WorkerRoutine) Close() error {
	if r.stop != nil {
		if err := r.stop(); err != nil {
			return fmt.Errorf("failed to stop %s: %w", r.name, err)
		}
	}
	return nil
}

// SetupInterruptHandler sets up signal handlers for graceful shutdown
// Returns a context that will be cancelled when an interrupt signal is received
func SetupInterruptHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	return ctx, cancel
}

// DetectAndKillHang monitors a context and detects if operations are hanging
// It will cancel the context if the timeout is exceeded
func DetectAndKillHang(ctx context.Context, timeout time.Duration, operationName string) (context.Context, context.CancelFunc) {
	hangCtx, hangCancel := context.WithTimeout(ctx, timeout)

	go func() {
		<-hangCtx.Done()
		if hangCtx.Err() == context.DeadlineExceeded {
			log.Printf("⚠️  HANG DETECTED: %s exceeded timeout of %v, cancelling...", operationName, timeout)
		}
	}()

	return hangCtx, hangCancel
}
