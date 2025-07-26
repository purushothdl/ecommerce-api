// internal/shared/tasks/client.go
package tasks

import (
	"context"
	"fmt"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
)

// NewClient creates and returns a new Google Cloud Tasks client.
func NewClient(ctx context.Context) (*cloudtasks.Client, error) {
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cloud tasks client: %w", err)
	}
	return client, nil
}