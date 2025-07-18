// internal/server/server.go
package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/domain"
)

type Server struct {
	config     *configs.Config
	logger     *slog.Logger
	router     *chi.Mux
	userService domain.UserService
	authService domain.AuthService
}

// New creates and initializes a new Server instance.
func New(
	config *configs.Config,
	logger *slog.Logger,
	userService domain.UserService,
	authService domain.AuthService,
) *Server {
	s := &Server{
		config:      config,
		logger:      logger,
		router:      chi.NewMux(),
		userService: userService,
		authService: authService,
	}

	s.registerRoutes()
	return s
}

// Router returns the server's HTTP handler.
func (s *Server) Router() http.Handler {
	return s.router
}