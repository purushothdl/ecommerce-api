// internal/server/server.go
package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/configs"
)

type Server struct {
	config  *configs.Config
	router  *chi.Mux
	db      *sql.DB
	logger  *log.Logger
}

func New(config *configs.Config, db *sql.DB, logger *log.Logger) *Server {
	s := &Server{
		config:  config,
		router:  chi.NewMux(),
		db:      db,
		logger:  logger,
	}
	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.router,
	}

	// We'll call RegisterRoutes here to set up all the handlers.
	s.RegisterRoutes()

	return server.ListenAndServe()
}