package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/database"
	"github.com/purushothdl/ecommerce-api/internal/server"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

type application struct {
	config      *configs.Config
	logger      *log.Logger
	userService user.Service
	authService auth.Service
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	cfg := configs.LoadConfig()

	// Initialize database connection
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Dependency injection
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	authService := auth.NewService(userService)

	app := &application{
		config:      cfg,
		logger:      logger,
		userService: userService,
		authService: authService,
	}

	// Start server
	srv := server.New(app.config, app.logger, app.userService, app.authService)
	logger.Printf("Starting %s server on port %d", cfg.Env, cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), srv.Router()); err != nil {
		logger.Fatalf("could not start server: %v", err)
	}
}
