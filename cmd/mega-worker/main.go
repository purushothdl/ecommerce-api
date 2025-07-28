// cmd/mega-worker/main.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/shared/tasks"
	apiclient "github.com/purushothdl/ecommerce-api/pkg/api-client"
	"github.com/purushothdl/ecommerce-api/workers/delivery"
	"github.com/purushothdl/ecommerce-api/workers/notification"
	"github.com/purushothdl/ecommerce-api/workers/shipping"
	"github.com/purushothdl/ecommerce-api/workers/warehouse"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := LoadWorkerConfig("worker.env")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize Cloud Tasks Client
	tasksClient, err := tasks.NewClient(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create tasks client: %w", err)
	}
	defer tasksClient.Close()

	taskCreatorCfg := tasks.TaskCreatorConfig{
		ProjectID:      cfg.GCPProjectID,
		LocationID:     cfg.GCPLocationID,
		QueueID:        cfg.GCPQueueID,
		WorkerURL:      cfg.WorkerURL,
		ServiceAccount: cfg.WorkerServiceAccount,
	}

	taskCreator := tasks.NewTaskCreator(tasksClient, taskCreatorCfg, logger)

	// Initialize our new API Client (for calling back to the main API)
	apiClient, err := apiclient.NewClient(context.Background(), cfg.ApiURL, logger)
	if err != nil {
		return fmt.Errorf("failed to create api client: %w", err)
	}

    // Initialize Template Service
    templateService, err := notification.NewTemplateService()
	if err != nil {
		return fmt.Errorf("failed to create template service: %w", err)
	}
    

	emailService := notification.NewEmailService(cfg.ResendAPIKey, cfg.ResendFromEmail, cfg.ResendFromName, logger)
	
	// Initialize handlers
	wh := warehouse.NewWarehouseHandler(logger, taskCreator, apiClient, cfg.WarehouseProcessingTime)
	sh := shipping.NewShippingHandler(logger, taskCreator, apiClient, cfg.ShippingProcessingTime)
	dh := delivery.NewDeliveryHandler(logger, taskCreator, apiClient, cfg.DeliveryProcessingTime)
	nh := notification.NewNotificationHandler(logger, emailService, templateService)

	// Setup router
	r := chi.NewRouter()
	r.Post("/handle/order-created", wh.HandleOrderCreated)
	r.Post("/handle/order-packed", sh.HandleOrderPacked)
	r.Post("/handle/order-shipped", dh.HandleOrderShipped)
	r.Post("/handle/notification-request", nh.HandleNotificationRequest)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Mega-worker is running."))
	})

	// Setup and start the server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("starting mega-worker server", "port", cfg.Port)
	return server.ListenAndServe()
}