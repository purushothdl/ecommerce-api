// internal/server/routes.go
package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/admin"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/shared/middleware"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

func (s *Server) registerRoutes() {
	isProduction := s.config.Env == "production"

	userHandler := user.NewHandler(s.userService, s.authService, s.logger)
	authHandler := auth.NewHandler(s.authService, s.config.JWT.Secret, isProduction, s.logger)
	adminHandler := admin.NewHandler(s.adminService, s.logger)

	// API versioning
	s.router.Route("/api/v1", func(r chi.Router) {
		s.registerV1Routes(r, userHandler, authHandler, adminHandler)
	})
}

func (s *Server) registerV1Routes(r chi.Router, userHandler *user.Handler, authHandler *auth.Handler, adminHandler *admin.Handler) {
	// Auth routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.TimeoutMiddleware(s.config.Timeouts.Auth))
		r.Post("/auth/login", authHandler.HandleLogin)
		r.Post("/auth/refresh", authHandler.HandleRefreshToken)
		r.Post("/auth/logout", authHandler.HandleLogout)
	})

	// User registration routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.TimeoutMiddleware(s.config.Timeouts.UserOps))
		r.Post("/users", userHandler.HandleRegister)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(s.config.JWT.Secret))
		r.Use(middleware.TimeoutMiddleware(s.config.Timeouts.Protected))

		// User profile routes
		r.Get("/users/profile", userHandler.HandleGetProfile)
		r.Put("/users/profile", userHandler.HandleUpdateProfile)
		r.Put("/users/password", userHandler.HandleChangePassword)
		r.Delete("/users/account", userHandler.HandleDeleteAccount)

		// Session management routes
		r.Get("/auth/sessions", authHandler.HandleGetSessions)
		r.Delete("/auth/sessions", authHandler.HandleLogoutAllDevices)
		r.Delete("/auth/sessions/{sessionId}", authHandler.HandleLogoutSpecificDevice)
	})

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(s.config.JWT.Secret))
		r.Use(middleware.AdminMiddleware)
		r.Use(middleware.TimeoutMiddleware(s.config.Timeouts.Protected))

		// User management routes
		r.Get("/admin/users", adminHandler.HandleListUsers)
		r.Post("/admin/users", adminHandler.HandleCreateUser)
		r.Put("/admin/users/{userId}", adminHandler.HandleUpdateUser)
		r.Delete("/admin/users/{userId}", adminHandler.HandleDeleteUser)
	})
}
