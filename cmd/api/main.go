// cmd/api/main.go
package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	cfg := configs.LoadConfig()

	db, err := sql.Open("pgx", cfg.DB.DSN)
	if err != nil {
		logger.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		logger.Fatalf("database not responding: %v", err)
	}
	logger.Println("Database connection established.")

	// Create a new instance of our server.
	// We need to pass it the dependencies it needs.
	// Let's modify the server struct and New function to accept these.

	srv := server.New(cfg, db, logger)

	// Start the server.
	logger.Printf("Starting %s server on port %d", cfg.Env, cfg.Port)
	if err := srv.Start(); err != nil {
		logger.Fatalf("could not start server: %v", err)
	}
}
