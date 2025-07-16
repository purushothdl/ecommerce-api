// internal/server/routes.go
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/user"
	"github.com/purushothdl/ecommerce-api/pkg/response"
)

func (s *Server) RegisterRoutes() {
	// For dependency injection into handlers
	userRepo := user.NewRepository(s.db) 
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(userService, s.config.JWT.Secret)

	// Public routes
	s.router.Post("/users", userHandler.HandleRegister)
	s.router.Post("/login", authHandler.HandleCreateToken)

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		r.Get("/profile", s.handleGetProfile)
	})
}

// handleGetProfile is a simple protected handler to test our middleware.
func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user from the context.
	// The authMiddleware guarantees this value will be present.
	userCtx, ok := r.Context().Value(auth.UserContextKey).(struct {
		ID   int64
		Role string
	})
	
	if !ok {
		// This should theoretically never happen if middleware is applied correctly.
		response.Error(w, http.StatusInternalServerError, "error retrieving user from context")
		return
	}

	// We can now use the user's ID and role for our business logic.
	response.JSON(w, http.StatusOK, map[string]any{
		"message": "Welcome to your protected profile!",
		"user_id": userCtx.ID,
		"role":    userCtx.Role,
	})
}