package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/pphelan007/davidAI/internal/config"
	"github.com/pphelan007/davidAI/internal/temporal/workflows"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get file path from command line args or use default
	filePath := "data/sine440.wav"
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.Address,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// Prepare workflow input
	workflowInput := workflows.AudioProcessingWorkflowInput{
		FilePath: filePath,
	}

	// Start workflow execution
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("audio-processing-%d", time.Now().Unix()),
		TaskQueue: cfg.Temporal.TaskQueue,
	}

	log.Printf("Starting AudioProcessingWorkflow with file: %s", filePath)
	log.Printf("Workflow ID: %s", workflowOptions.ID)
	log.Printf("Task Queue: %s", cfg.Temporal.TaskQueue)

	workflowRun, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.AudioProcessingWorkflow, workflowInput)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	log.Printf("Workflow started! Workflow ID: %s, Run ID: %s", workflowRun.GetID(), workflowRun.GetRunID())

	// Wait for workflow to complete
	var result workflows.AudioProcessingWorkflowOutput
	err = workflowRun.Get(context.Background(), &result)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	// Print results
	log.Println("✅ Workflow completed successfully!")
	log.Printf("Ingested Asset ID: %s", result.IngestedAsset.AssetID)
	log.Printf("Ingested Asset Path: %s", result.IngestedAsset.FilePath)
	log.Printf("Ingested Asset Duration: %.2f seconds", result.IngestedAsset.Metadata.Duration)
	log.Printf("Trimmed Output Path: %s", result.TrimmedOutput.OutputPath)
	log.Printf("Was Trimmed: %v", result.TrimmedOutput.WasTrimmed)
	log.Printf("No Op: %v", result.TrimmedOutput.NoOp)
}
