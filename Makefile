# Makefile
.PHONY: migrateup migratedown run

# Load environment variables from .env file
include .env
export

# Use the DB_DSN directly if it's already a valid PostgreSQL URL
MIGRATE_DSN := $(DB_DSN)

migrateup:
	@echo "Running migrations up..."
	migrate -database $(MIGRATE_DSN) -path migrations up

migratedown:
	@echo "Running migrations down..."
	migrate -database $(MIGRATE_DSN) -path migrations down

run:
	@echo "Starting server..."
	go run ./cmd/api/main.go

hashtest:
	@echo "Running hash test..."
	go run ./hashtest/main.go