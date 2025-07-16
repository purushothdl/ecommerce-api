// configs/config.go
package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN string
	}
	JWT struct {
		Secret string
	}
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
	// Load .env file. In production, these vars would be set directly.
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	var cfg Config

	cfg.Env = getEnv("ENV", "development")
	cfg.Port = getEnvAsInt("PORT", 8080)
	cfg.DB.DSN = getEnv("DB_DSN", "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable")

    cfg.JWT.Secret = getEnv("JWT_SECRET", "default-super-secret-key") 
    return &cfg
}

// Helper function to read an environment variable or return a default.
func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        log.Printf("Using %s=%s", key, value)
        return value
    }
    log.Printf("Using default for %s=%s", key, fallback)
    return fallback
}

// Helper function to read an environment variable as an integer or return a default.
func getEnvAsInt(key string, fallback int) int {
    if valueStr, exists := os.LookupEnv(key); exists {
        if value, err := strconv.Atoi(valueStr); err == nil {
            if value > 0 && value < 65536 { // Valid port range
                return value
            }
            log.Printf("Invalid port %d, using default %d", value, fallback)
        }
    }
    return fallback
}