// internal/server/server.go
package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

type Server struct {
	config      *configs.Config
	logger      *slog.Logger
	router      *chi.Mux
	userService user.Service
	authService auth.Service
}

func New(config *configs.Config, logger *slog.Logger, userService user.Service, authService auth.Service) *Server {
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

// Router returns the server's router.
func (s *Server) Router() http.Handler {
	return s.router
}