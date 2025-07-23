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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/address"
	"github.com/purushothdl/ecommerce-api/internal/admin"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/cart"
	"github.com/purushothdl/ecommerce-api/internal/category"
	"github.com/purushothdl/ecommerce-api/internal/database"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/product"
	"github.com/purushothdl/ecommerce-api/internal/server"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

type application struct {
	config          *configs.Config
	logger          *slog.Logger
	userService     domain.UserService
	authService     domain.AuthService
	adminService    domain.AdminService
	productService  domain.ProductService
	categoryService domain.CategoryService
	store 			domain.Store
	cartService     domain.CartService
	addressService  domain.AddressService
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)) 
	
	// Load configuration
	cfg, err := configs.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize database
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	logger.Info("database connection established", 
		"max_open_conns", cfg.DB.MaxOpenConns,
		"max_idle_conns", cfg.DB.MaxIdleConns,
	)

	// Initialize store for transactions
	store := database.NewStore(db)

	// Setup repositories (implement domain interfaces)
	userRepo := user.NewUserRepository(db)
	authRepo := auth.NewAuthRepository(db)
	categoryRepo := category.NewCategoryRepository(db)
	productRepo := product.NewProductRepository(db)
	cartRepo := cart.NewCartRepository(db)
	addressRepo := address.NewAddressRepository(db)

	// Setup services (implement domain interfaces)
	cartService := cart.NewCartService(cartRepo, productRepo, store, logger)
	authService := auth.NewAuthService(authRepo, userRepo, cartService, cfg.JWT.Secret, logger)
	userService := user.NewUserService(userRepo, authService, cartService, logger)	
	adminService := admin.NewAdminService(userRepo, logger)
	categoryService := category.NewCategoryService(categoryRepo, logger)
	productService := product.NewProductService(productRepo, logger)
	addressService := address.NewAddressService(addressRepo, store, logger)

	app := &application{
		config:          cfg,
		logger:          logger,
		userService:     userService,
		authService:     authService,
		adminService:    adminService,
		productService:  productService,
		categoryService: categoryService,
		store: 			 store,
		cartService:     cartService,
		addressService:  addressService,
	}

	// Start server
	return app.startServer()
}

func (app *application) startServer() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler: server.New(
			app.config, app.logger, app.userService, app.authService,
			app.adminService, app.productService, app.categoryService,
			app.cartService, app.store, app.addressService,
		).Router(),
		ReadTimeout:  app.config.Server.ReadTimeout,
		WriteTimeout: app.config.Server.WriteTimeout,
		IdleTimeout:  app.config.Server.IdleTimeout,
	}

	// Graceful shutdown
	shutdownComplete := make(chan struct{})
	
	if app.config.Server.GracefulShutdown {
		go app.handleShutdown(srv, shutdownComplete)
	}

	app.logger.Info("starting server",
		"env", app.config.Env,
		"port", app.config.Port,
		"read_timeout", app.config.Server.ReadTimeout,
		"write_timeout", app.config.Server.WriteTimeout,
	)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed to start: %w", err)
	}

	if app.config.Server.GracefulShutdown {
		<-shutdownComplete
		app.logger.Info("server shutdown complete")
	}

	return nil
}

func (app *application) handleShutdown(srv *http.Server, shutdownComplete chan struct{}) {
	defer close(shutdownComplete)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-quit
	app.logger.Info("received shutdown signal", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), app.config.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.logger.Error("forced server shutdown", "error", err)
	}
}
