// configs/config.go
package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/purushothdl/ecommerce-api/internal/shared/tasks"
)

// Main configuration struct
type Config struct {
	Env             string
	Port            int
	DB              DBConfig
	JWT             JWTConfig
	Server          ServerConfig
	Timeouts        TimeoutConfig
	CORS            CORSConfig
	Stripe          StripeConfig
	ApiURL          string
	OrderFinancials *OrderFinancialsConfig
	GCTasks         tasks.TaskCreatorConfig
}

// Database configuration
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// JWT authentication configuration
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// Server operation configuration
type ServerConfig struct {
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	ShutdownTimeout  time.Duration
	GracefulShutdown bool
}

// Timeout configuration for various operations
type TimeoutConfig struct {
	Auth      time.Duration
	UserOps   time.Duration
	Protected time.Duration
	Database  time.Duration
}

// CORS configuration for API access control
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// Order financials configuration
type OrderFinancialsConfig struct {
	OrderShippingCost    float64
	OrderTaxRate         float64
	OrderDiscountAmount  float64
}

// Stripe payment configuration
type StripeConfig struct {
	SecretKey      string
	WebhookSecret  string
}

func LoadConfig(path string) (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	if err := godotenv.Load(path); err != nil && os.Getenv("ENV") != "production" {
		// Only log in development
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	cfg := &Config{
		Env:  getEnv("ENV", "development"),
		Port: getEnvAsInt("PORT", 8080),

		DB: DBConfig{
			DSN:             getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},

		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", "default-secret-change-in-production"),
			AccessTokenDuration:  getEnvAsDuration("JWT_ACCESS_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDuration("JWT_REFRESH_DURATION", 7*24*time.Hour),
		},

		Server: ServerConfig{
			ReadTimeout:      getEnvAsDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:     getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:      getEnvAsDuration("SERVER_IDLE_TIMEOUT", time.Minute),
			ShutdownTimeout:  getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			GracefulShutdown: getEnvAsBool("SERVER_GRACEFUL_SHUTDOWN", true),
		},

		Timeouts: TimeoutConfig{
			Auth:      getEnvAsDuration("TIMEOUT_AUTH", 10*time.Second),
			UserOps:   getEnvAsDuration("TIMEOUT_USER_OPS", 15*time.Second),
			Protected: getEnvAsDuration("TIMEOUT_PROTECTED", 8*time.Second),
			Database:  getEnvAsDuration("TIMEOUT_DATABASE", 5*time.Second),
		},

		CORS: CORSConfig{
			AllowOrigins:     strings.Split(getEnv("CORS_ALLOW_ORIGINS", "*"), ","),
			AllowMethods:     strings.Split(getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH"), ","),
			AllowHeaders:     strings.Split(getEnv("CORS_ALLOW_HEADERS", "Accept,Authorization,Content-Type,X-CSRF-Token"), ","),
			ExposeHeaders:    strings.Split(getEnv("CORS_EXPOSE_HEADERS", ""), ","),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 86400),
		},

		Stripe: StripeConfig{
			SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},

		ApiURL: getEnv("ECOMMERCE_API_URL", ""),

		OrderFinancials: &OrderFinancialsConfig{
			OrderShippingCost:    getEnvAsFloat64("ORDER_SHIPPING_COST", 50.00),
			OrderTaxRate:         getEnvAsFloat64("ORDER_TAX_RATE", 0.18),
			OrderDiscountAmount:  getEnvAsFloat64("ORDER_DISCOUNT_AMOUNT", 0.00),
		},

		GCTasks: tasks.TaskCreatorConfig{
			ProjectID:      getEnv("GCP_PROJECT_ID", ""),
			LocationID:     getEnv("GCP_TASKS_LOCATION_ID", ""),
			QueueID:        getEnv("GCP_TASKS_QUEUE_ID", ""),
			WorkerURL:      getEnv("MEGA_WORKER_URL", ""),
			ServiceAccount: getEnv("MEGA_WORKER_SA_EMAIL", ""),
		},

	}

	// Validate critical config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate ensures configuration values are valid
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.JWT.Secret == "default-secret-change-in-production" && c.Env == "production" {
		return fmt.Errorf("default JWT secret cannot be used in production")
	}

	if c.DB.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	return nil
}

// Helper functions
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return fallback
}

func getEnvAsFloat64(key string, fallback float64) float64 {
    if valueStr, exists := os.LookupEnv(key); exists {
        if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
            return value
        }
    }
    return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return fallback
}
