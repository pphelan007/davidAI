// Package activities contains Temporal activity definitions and client.
package activities

import "go.temporal.io/sdk/worker"

// RegisterActivities registers all activities with the given Temporal worker
func RegisterActivities(w worker.Worker, activitiesClient *ActivitiesClient) {
	// Register audio processing activities
	w.RegisterActivity(activitiesClient.IngestRawAudio)
	w.RegisterActivity(activitiesClient.TrimSilence)
}
