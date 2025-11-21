// Package workflows contains Temporal workflow definitions.
package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/pphelan007/davidAI/internal/temporal/activities"
)

// AudioProcessingWorkflowInput is the input for the AudioProcessingWorkflow
type AudioProcessingWorkflowInput struct {
	FilePath string `json:"file_path"`
}

// AudioProcessingWorkflowOutput is the output from the AudioProcessingWorkflow
type AudioProcessingWorkflowOutput struct {
	IngestedAsset activities.AssetInfo         `json:"ingested_asset"`
	TrimmedOutput activities.TrimSilenceOutput `json:"trimmed_output"`
}

// AudioProcessingWorkflow is a simple workflow that ingests raw audio and trims silence
func AudioProcessingWorkflow(ctx workflow.Context, input AudioProcessingWorkflowInput) (*AudioProcessingWorkflowOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Ingest raw audio from the data folder
	var ingestOutput *activities.IngestRawAudioOutput
	err := workflow.ExecuteActivity(ctx, "IngestRawAudio", activities.IngestRawAudioInput{
		FilePath: input.FilePath,
	}).Get(ctx, &ingestOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest raw audio: %w", err)
	}

	// Step 2: Trim silence (which internally uses findNonSilentRange)
	var trimOutput *activities.TrimSilenceOutput
	err = workflow.ExecuteActivity(ctx, "TrimSilence", activities.TrimSilenceInput{
		AssetID:            ingestOutput.Asset.AssetID,
		SourcePath:         ingestOutput.Asset.FilePath,
		SilenceThreshold:   0.01, // 1% threshold
		MinSilenceDuration: 0.1,  // 100ms minimum silence
	}).Get(ctx, &trimOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to trim silence: %w", err)
	}

	return &AudioProcessingWorkflowOutput{
		IngestedAsset: ingestOutput.Asset,
		TrimmedOutput: *trimOutput,
	}, nil
}

// RegisterWorkflows registers all workflows with the given Temporal worker
func RegisterWorkflows(w worker.Worker) {
	w.RegisterWorkflow(AudioProcessingWorkflow)
}
