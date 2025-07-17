# E-Commerce API

## Project Structure

### Handler-Service-Repository Pattern
Our application follows a clean architecture pattern:

1. **Handlers**:
   - Handle HTTP requests/responses
   - Perform input validation
   - Call service layer
   - Located in `internal/{domain}/handler.go`

2. **Services**:
   - Contain business logic
   - Interface with repositories
   - Located in `internal/{domain}/service.go`

3. **Repositories**:
   - Handle database operations
   - Located in `internal/{domain}/repository.go`

### Error Handling
- Shared errors are defined in `pkg/errors`
- Domain-specific errors are defined in their respective packages

### Middleware
- Authentication middleware in `internal/server/middleware.go`
- Adds user context to requests

## Getting Started

### Prerequisites
- Go 1.21+ ([install guide](https://go.dev/doc/install))
- PostgreSQL 15+ ([install guide](https://www.postgresql.org/download/))
- Migrate CLI ([install guide](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate))

### Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/purushothdl/e-commerce-api.git
   cd e-commerce-api
   ```

2. Configure environment:
   ```bash
   cp .env.example .env
   ```
   Edit `.env` with your PostgreSQL credentials:
   ```env
   DB_DSN=postgres://user:password@localhost:5432/ecommerce?sslmode=disable
   JWT_SECRET=your-secret-key
   ```

### Database Setup
#### With Make (recommended):
```bash
make migrateup  # Applies migrations
make migratedown  # Rolls back migrations (if needed)
```

#### Without Make:
```bash
# Apply migrations
migrate -database "$DB_DSN" -path migrations up

# Rollback migrations
migrate -database "$DB_DSN" -path migrations down
```

### Running the Server
#### With Make:
```bash
make run
```

#### Without Make:
```bash
go run ./cmd/api/main.go
```

The server will start at `http://localhost:8080`

## API Endpoints
| Method | Endpoint       | Description                |
|--------|----------------|----------------------------|
| POST   | /users         | Register new user          |
| POST   | /login         | Authenticate user          |
| GET    | /profile       | Get user profile (protected)|

## Testing
Run unit tests:
```bash
go test ./...
```

## Documentation
- Architecture decisions: `docs/architecture/decisions/`
- API specs: `docs/api-specs/openapi.yaml` (generate with `swag init`)

## Troubleshooting
- **Database connection issues**: Verify `DB_DSN` in `.env` matches your PostgreSQL credentials
- **Migration errors**: Ensure PostgreSQL is running and the database exists
- **JWT errors**: Confirm `JWT_SECRET` in `.env` is consistent