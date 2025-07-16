# Makefile
.PHONY: migrateup migratedown

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