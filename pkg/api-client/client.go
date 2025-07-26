// pkg/api-client/client.go
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// Client is a client for interacting with the main ecommerce API.
type Client struct {
	apiURL     string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient creates a new, authenticated API client.
// It sets up an http.Client that automatically adds a Google OIDC token to every request.
func NewClient(ctx context.Context, audience string, logger *slog.Logger) (*Client, error) {
	if audience == "" {
		return nil, fmt.Errorf("API client audience (URL) cannot be empty")
	}

	tokenSource, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("failed to create idtoken source for api client: %w", err)
	}

	// Create an http.Client that uses the OIDC token source.
	httpClient := oauth2.NewClient(ctx, tokenSource)

	return &Client{
		apiURL:     audience,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// UpdateOrderStatus calls the internal API endpoint to update an order's status.
func (c *Client) UpdateOrderStatus(ctx context.Context, orderID int64, payload dto.UpdateOrderStatusRequest) error {
	c.logger.Info("Sending API status update", "order_id", orderID, "payload", payload)

	// The payload is now the flexible DTO struct itself.
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("apiclient: failed to marshal status payload: %w", err)
	}

	// Construct the full URL for the internal endpoint
	url := fmt.Sprintf("%s/api/v1/internal/orders/%d/status", c.apiURL, orderID)

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("apiclient: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request using the authenticated http.Client
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("apiclient: failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("apiclient: status update failed with status code: %d", resp.StatusCode)
	}

	c.logger.Info("Successfully sent API status update", "order_id", orderID)
	return nil
}