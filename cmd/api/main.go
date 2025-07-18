// cmd/api/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/database"
	"github.com/purushothdl/ecommerce-api/internal/server"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

type application struct {
	config      *configs.Config
	logger      *slog.Logger
	userService user.Service
	authService auth.Service
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := configs.LoadConfig()

	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	authService := auth.NewService(userRepo)

	app := &application{
		config:      cfg,
		logger:      logger,
		userService: userService,
		authService: authService,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      server.New(app.config, app.logger, app.userService, app.authService).Router(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	// Channel to signal that the shutdown is complete.
	shutdownComplete := make(chan struct{})

	// Goroutine for graceful shutdown
	go func() {
		// Signal to the main goroutine that shutdown is complete.
		defer close(shutdownComplete)

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failed", "error", err)
		}
	}()

	logger.Info("starting server", "env", cfg.Env, "port", cfg.Port)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("could not start server", "error", err)
		os.Exit(1)
	}

	// Block here and wait for the shutdown goroutine to complete.
	<-shutdownComplete
	logger.Info("server exited gracefully")
}