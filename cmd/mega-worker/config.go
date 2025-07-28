// cmd/mega-worker/config.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type WorkerConfig struct {
	// Server Configuration
	Port string
	Env  string
	
	// API Configuration
	ApiURL string
	
	// Google Cloud Configuration
	GCPProjectID         string
	GCPLocationID        string
	GCPQueueID           string
	WorkerURL            string
	WorkerServiceAccount string
	
	// Email Configuration
	ResendAPIKey    string
	ResendFromEmail string
	ResendFromName  string
	
	// Processing Times
	WarehouseProcessingTime time.Duration
	ShippingProcessingTime  time.Duration
	DeliveryProcessingTime  time.Duration
}

func LoadWorkerConfig(path string) (*WorkerConfig, error) {
	// Not needed in prodution
	if err := godotenv.Load(path); err != nil && os.Getenv("ENV") != "production" {
		fmt.Printf("Warning: .env file not found at %s: %v\n", path, err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	return &WorkerConfig{
		Port:                 port,
		Env:                  os.Getenv("ENV"),
		ApiURL:               os.Getenv("ECOMMERCE_API_URL"),
		GCPProjectID:         os.Getenv("GCP_PROJECT_ID"),
		GCPLocationID:        os.Getenv("GCP_TASKS_LOCATION_ID"),
		GCPQueueID:           os.Getenv("GCP_TASKS_QUEUE_ID"),
		WorkerURL:            os.Getenv("MEGA_WORKER_URL"),
		WorkerServiceAccount: os.Getenv("MEGA_WORKER_SA_EMAIL"),
		ResendAPIKey:         os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:      os.Getenv("RESEND_FROM_EMAIL"),
		ResendFromName:       os.Getenv("RESEND_FROM_NAME"),
	}, nil
}
