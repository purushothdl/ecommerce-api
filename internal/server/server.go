// internal/server/server.go
package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/shared/middleware"
)

type Server struct {
	config       *configs.Config
	logger       *slog.Logger
	router       *chi.Mux
	userService  domain.UserService
	authService  domain.AuthService
	adminService domain.AdminService
}

// New creates and initializes a new Server instance.
func New(
	config       *configs.Config,
	logger       *slog.Logger,
	userService  domain.UserService,
	authService  domain.AuthService,
	adminService domain.AdminService,
) *Server {
	s := &Server{
		config:       config,
		logger:       logger,
		router:       chi.NewMux(),
		userService:  userService,
		authService:  authService,
		adminService: adminService,
	}
	// Apply global middleware in order
	s.router.Use(middleware.RecoveryMiddleware(logger))
	s.router.Use(middleware.LoggingMiddleware(logger))
	s.router.Use(middleware.CORSMiddleware(config.CORS))

	s.registerRoutes()
	return s
}

// Router returns the server's HTTP handler.
func (s *Server) Router() http.Handler {
	return s.router
}
