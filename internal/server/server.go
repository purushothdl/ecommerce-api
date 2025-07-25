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
	config          *configs.Config
	logger          *slog.Logger
	router          *chi.Mux
	userService     domain.UserService
	authService     domain.AuthService
	adminService    domain.AdminService
	productService  domain.ProductService
	categoryService domain.CategoryService
	cartService     domain.CartService
	store 			domain.Store
	addressService  domain.AddressService
	orderService    domain.OrderService
	paymentService  domain.PaymentService
	isProduction    bool 
}

// New creates and initializes a new Server instance.
func New(
	config          *configs.Config,
	logger          *slog.Logger,
	userService     domain.UserService,
	authService     domain.AuthService,
	adminService    domain.AdminService,
	productService  domain.ProductService,
	categoryService domain.CategoryService,
	cartService     domain.CartService,	
	store 			domain.Store,
	addressService  domain.AddressService,
	orderService    domain.OrderService,
	paymentService  domain.PaymentService,

) *Server {
	s := &Server{
		config:          config,
		logger:          logger,
		router:          chi.NewMux(),
		userService:     userService,
		authService:     authService,
		adminService:    adminService,
		productService:  productService,
		categoryService: categoryService,
		cartService:     cartService,
		store: 			 store,
		addressService:  addressService,
		orderService:    orderService,
		paymentService:  paymentService,
		isProduction:    config.Env == "production", 
	}

	// Apply global middleware in order
	s.router.Use(middleware.RecoveryMiddleware(logger))
	s.router.Use(middleware.LoggingMiddleware(logger))
	s.router.Use(middleware.ChiCors(s.config.CORS))

	s.registerRoutes()
	return s
}

// Router returns the server's HTTP handler.
func (s *Server) Router() http.Handler {
	return s.router
}
