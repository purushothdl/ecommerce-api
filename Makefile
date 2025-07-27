# Makefile
.PHONY: migrateup migratedown run hashtest seed test-payment run-worker

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

seed:
	@echo "Seeding database..."
	@go run ./cmd/seed/

test-payment:
	@echo "Starting test payment server..."
	python -m http.server --directory static 8005

run-workers:
	@echo "Starting mega worker..."
	go run ./cmd/mega-worker/

