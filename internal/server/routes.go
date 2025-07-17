// internal/server/routes.go
package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

func (s *Server) RegisterRoutes() {
	// Initialize dependencies
	userRepo := user.NewRepository(s.db)
	
	userService := user.NewService(userRepo)
	authService := auth.NewService(userService)

	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService, s.config.JWT.Secret)

	// Public routes
	s.router.Post("/users", userHandler.HandleRegister)
	s.router.Post("/login", authHandler.HandleLogin)

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Get("/profile", userHandler.HandleGetProfile)
	})
}
