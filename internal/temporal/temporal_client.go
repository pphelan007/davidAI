package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
)

type TemporalClient struct {
	client client.Client
}

// NewTemporalClient creates a new Temporal client connected to the local dev server
func NewTemporalClient(ctx context.Context, address, namespace string) (*TemporalClient, error) {
	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial Temporal server at %s: %w", address, err)
	}

	return &TemporalClient{
		client: c,
	}, nil
}

// GetClient returns the underlying Temporal client
func (t *TemporalClient) GetClient() client.Client {
	return t.client
}

// Close closes the Temporal client connection
func (t *TemporalClient) Close() error {
	t.client.Close()
	return nil
}
