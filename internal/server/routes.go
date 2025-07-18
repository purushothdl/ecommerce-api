// internal/server/routes.go
package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

func (s *Server) registerRoutes() {
	// Create handlers using the dependencies from the server struct.
	userHandler := user.NewHandler(s.userService)
	authHandler := auth.NewHandler(s.authService, s.config.JWT.Secret)

	// Public routes
	s.router.Post("/users", userHandler.HandleRegister)
	s.router.Post("/login", authHandler.HandleLogin)
	s.router.Post("/auth/refresh", authHandler.HandleRefreshToken)
	s.router.Post("/logout", authHandler.HandleLogout) 

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Get("/profile", userHandler.HandleGetProfile)
	})
}
