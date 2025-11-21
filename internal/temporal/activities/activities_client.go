package activities

import (
	"context"

	"go.temporal.io/sdk/client"
)

type ActivitiesClient struct {
	client client.Client
}

func NewActivitiesClient(ctx context.Context, client client.Client) *ActivitiesClient {
	return &ActivitiesClient{
		client: client,
	}
}
