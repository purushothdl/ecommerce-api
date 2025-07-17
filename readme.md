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
[... rest of your README ...]