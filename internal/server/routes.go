// internal/server/routes.go
package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

func (s *Server) registerRoutes() {
	userHandler := user.NewHandler(s.userService)
	authHandler := auth.NewHandler(s.authService, s.config.JWT.Secret)

	// API versioning
	s.router.Route("/api/v1", func(r chi.Router) {
		s.registerV1Routes(r, userHandler, authHandler)
	})
}

func (s *Server) registerV1Routes(r chi.Router, userHandler *user.Handler, authHandler *auth.Handler) {
	// Auth routes
	r.Group(func(r chi.Router) {
		r.Use(s.timeoutMiddleware(s.config.Timeouts.Auth))
		r.Post("/auth/login", authHandler.HandleLogin)
		r.Post("/auth/refresh", authHandler.HandleRefreshToken)
		r.Post("/auth/logout", authHandler.HandleLogout)
	})

	// User registration routes
	r.Group(func(r chi.Router) {
		r.Use(s.timeoutMiddleware(s.config.Timeouts.UserOps))
		r.Post("/users", userHandler.HandleRegister)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Use(s.timeoutMiddleware(s.config.Timeouts.Protected))
		r.Get("/users/profile", userHandler.HandleGetProfile)
	})
}
