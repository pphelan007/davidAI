// Package workflows contains Temporal workflow definitions.
package workflows

import "go.temporal.io/sdk/worker"

// RegisterWorkflows registers all workflows with the given Temporal worker
func RegisterWorkflows(w worker.Worker) {
	// Register workflows here
	// Example: w.RegisterWorkflow(DataProcessingWorkflow)

	// Placeholder for future workflow registration
	_ = w
}
