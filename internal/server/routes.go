// internal/server/routes.go
package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/admin"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/cart"
	"github.com/purushothdl/ecommerce-api/internal/product"
	"github.com/purushothdl/ecommerce-api/internal/shared/middleware"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

func (s *Server) registerRoutes() {
	userHandler := user.NewHandler(s.userService, s.authService, s.cartService, s.config.JWT.Secret, s.isProduction, s.logger)
	authHandler := auth.NewHandler(s.authService, s.cartService, s.config.JWT.Secret, s.isProduction, s.logger)
	adminHandler := admin.NewHandler(s.adminService, s.logger)
	productHandler := product.NewHandler(s.productService, s.categoryService, s.logger)
	cartHandler := cart.NewHandler(s.cartService, s.logger)

	// API versioning
	s.router.Route("/api/v1", func(r chi.Router) {
		s.registerV1Routes(r, userHandler, authHandler, adminHandler, productHandler, cartHandler)
	})
}

func (s *Server) registerV1Routes(r chi.Router, userHandler *user.Handler, authHandler *auth.Handler, adminHandler *admin.Handler, productHandler *product.Handler, cartHandler *cart.Handler) {
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

	// Public product and category routes (no authentication required)
	r.Group(func(r chi.Router) {
        r.Get("/products", productHandler.HandleListProducts)
        r.Get("/products/{productId}", productHandler.HandleGetProduct)
        r.Get("/categories", productHandler.HandleListCategories)
    })

	// Cart routes with cart middleware for session/user cart management
	r.Group(func(r chi.Router) {
		r.Use(middleware.OptionalAuthMiddleware(s.config.JWT.Secret)) 
        r.Use(middleware.CartMiddleware(s.cartService, s.isProduction))
        
        r.Get("/cart", cartHandler.HandleGetCart)
        r.Post("/cart/items", cartHandler.HandleAddItem)
        
        // These routes operate on a specific product within the cart
        r.Patch("/cart/items/{productId}", cartHandler.HandleUpdateItem)
        r.Delete("/cart/items/{productId}", cartHandler.HandleRemoveItem)
    })

}
