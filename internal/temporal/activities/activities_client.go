package activities

import (
	"context"

	"go.temporal.io/sdk/client"

	"github.com/pphelan007/davidAI/internal/database"
)

type ActivitiesClient struct {
	client   client.Client
	dbClient *database.Client
}

func NewActivitiesClient(ctx context.Context, temporalClient client.Client, dbClient *database.Client) *ActivitiesClient {
	return &ActivitiesClient{
		client:   temporalClient,
		dbClient: dbClient,
	}
}
