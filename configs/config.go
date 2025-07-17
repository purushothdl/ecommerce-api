// configs/config.go
package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	DSN string
}

type JWTConfig struct {
	Secret string
}

type Config struct {
	Port int
	Env  string
	DB   DBConfig
	JWT  JWTConfig
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Env:  getEnv("ENV", "development"),
		Port: getEnvAsInt("PORT", 8080),
		DB:   DBConfig{DSN: getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")},
		JWT:  JWTConfig{Secret: getEnv("JWT_SECRET", "default-secret")},
	}
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
