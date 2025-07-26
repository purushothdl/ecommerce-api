// internal/shared/tasks/creator.go
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"

)

// TaskCreatorConfig holds the configuration required to create tasks.
type TaskCreatorConfig struct {
	ProjectID      string
	LocationID     string
	QueueID        string
	WorkerURL      string
	ServiceAccount string
}

// TaskCreator is responsible for creating and enqueuing tasks.
type TaskCreator struct {
	client *cloudtasks.Client
	config TaskCreatorConfig
	logger *slog.Logger
}

// NewTaskCreator returns a new TaskCreator.
func NewTaskCreator(client *cloudtasks.Client, config TaskCreatorConfig, logger *slog.Logger) *TaskCreator {
	return &TaskCreator{
		client: client,
		config: config,
		logger: logger,
	}
}

// createTask is a generic helper to construct and create a task.
func (tc *TaskCreator) createTask(ctx context.Context, handlerPath string, payload any) error {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tc.config.ProjectID, tc.config.LocationID, tc.config.QueueID)

	// Marshal the event payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	req := &cloudtaskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &cloudtaskspb.Task{
			MessageType: &cloudtaskspb.Task_HttpRequest{
				HttpRequest: &cloudtaskspb.HttpRequest{
					HttpMethod: cloudtaskspb.HttpMethod_POST,
					Url:        tc.config.WorkerURL + handlerPath,
					Body:       jsonPayload,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					// Use OIDC token for secure authentication between services
					AuthorizationHeader: &cloudtaskspb.HttpRequest_OidcToken{
						OidcToken: &cloudtaskspb.OidcToken{
							ServiceAccountEmail: tc.config.ServiceAccount,
						},
					},
				},
			},
			// ScheduleTime: timestamppb.New(time.Now().Add(10 * time.Second)),
		},
	}

	createdTask, err := tc.client.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create cloud task: %w", err)
	}

	tc.logger.Info("Task created successfully", "task_name", createdTask.Name, "handler_path", handlerPath)
	return nil
}

// CreateFulfillmentTask is an example of a specific task creation method.
func (tc *TaskCreator) CreateFulfillmentTask(ctx context.Context, handlerPath string, payload any) error {
	return tc.createTask(ctx, handlerPath, payload)
}