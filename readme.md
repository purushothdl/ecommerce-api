# Go E-Commerce API

![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Chi Router](https://img.shields.io/badge/Chi-v5-FF6F61.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-4169E1.svg)
![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED.svg)

A robust and feature-rich backend for an e-commerce platform, built with Go. This project is a demonstration of modern backend development best practices, including a clean, layered architecture, secure authentication, and robust database management. It is designed to be scalable, maintainable, and easy to extend.

## Key Features

-   **Secure Authentication:**
    -   Stateful session management using JWT access tokens and secure, HttpOnly refresh tokens.
    -   Password hashing using `bcrypt`.
    -   Complete session control: view active sessions, log out from a specific device, or log out from all devices.
-   **User Management:**
    -   User registration and profile management (update details, change password).
    -   Secure account deletion.
-   **Administrator Role:**
    -   Role-based access control (RBAC) via middleware.
    -   Admin-only endpoints for complete User CRUD (Create, Read, Update, Delete) operations.
-   **Product & Category System:**
    -   List products with filtering (by category) and pagination.
    -   View detailed product information.
    -   List all product categories.
-   **Shopping Cart Logic:**
    -   Persistent carts for both authenticated and anonymous users.
    -   Seamlessly merges an anonymous user's cart to their account upon registration or login.
    -   Full cart item management (add, update quantity, remove).
-   **Database Seeding:**
    -   A powerful seeder script to populate the database with realistic product data from an external API (`dummyjson.com`) and create a default admin user.

## Architectural Highlights

This project was built with a strong emphasis on clean architecture and professional development patterns.

-   **Clean Architecture:** Follows the **Handler-Service-Repository** pattern to ensure a clear separation of concerns.
    -   **Handlers:** Manage HTTP request/response logic.
    -   **Services:** Contain the core business logic.
    -   **Repositories:** Abstract database interactions.
-   **Secure by Design:** The authentication flow is built on modern security standards to prevent common vulnerabilities like CSRF and XSS attacks related to token handling.
-   **Transactional Integrity:** Critical operations that involve multiple database writes (e.g., user registration with cart merging) are executed within a single **atomic transaction** to ensure data consistency.
-   **Structured Logging:** Utilizes the structured logging library `slog` for clear, context-rich application logs that are invaluable for debugging.
-   **Configuration Management:** All environment-specific settings (database connections, JWT secrets, server timeouts) are managed via a dedicated `configs` package and loaded from a `.env` file, ensuring no sensitive data is hardcoded.
-   **Robust Middleware:** A well-defined middleware chain handles logging, CORS, panic recovery, request timeouts, and authentication, keeping the handler logic clean and focused.

## Tech Stack

-   **Language:** Go
-   **Web Framework:** [Chi (v5)](https://github.com/go-chi/chi) - A lightweight, idiomatic and composable router for building Go HTTP services.
-   **Database:** PostgreSQL
-   **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate) - For database schema management.
-   **Authentication:** JWT & Refresh Tokens
-   **Environment:** godotenv

## Project Structure

The project follows the standard layout for Go applications, promoting scalability and maintainability.

```
.
├── cmd/                # Application entry points (main.go)
│   ├── api/            # Main API server
│   └── seed/           # Database seeder
├── configs/            # Configuration management
├── internal/           # Private application logic
│   ├── admin/          # Admin-specific domain
│   ├── auth/           # Authentication domain
│   ├── cart/           # Shopping cart domain
│   ├── product/        # Product domain
│   ├── user/           # User domain
│   ├── database/       # Database connection & store
│   ├── domain/         # Core domain types, interfaces
│   ├── models/         # Database models
│   ├── server/         # HTTP server setup & routing
│   └── shared/         # Shared code (middleware, context)
├── migrations/         # SQL database migrations
├── pkg/                # Public library code (errors, response helpers)
└── ...
```

## Getting Started

Follow these instructions to get the project up and running on your local machine.

### Prerequisites

-   Go 1.21+
-   PostgreSQL 15+
-   `make` (optional, for convenience)
-   [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### 1. Clone the Repository

```bash
git clone https://github.com/purushothdl/ecommerce-api.git
cd ecommerce-api
```

### 2. Configure Environment

Copy the example `.env` file and update it with your local PostgreSQL details.

```bash
cp .env.example .env
```

Your `.env` file should look like this:

```env
# Server Configuration
PORT=8080
ENV=development

# Database Configuration (Update with your credentials)
DB_DSN=postgres://youruser:yourpassword@localhost:5432/ecommerce?sslmode=disable

# JWT Configuration
JWT_SECRET=a-very-strong-and-secret-key-that-is-at-least-32-bytes-long

# Admin user for seeder (Change this!)
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=supersecretpassword
```

**Note:** Ensure you have created a database named `ecommerce` in PostgreSQL.

### 3. Run Database Migrations

This will set up all the necessary tables in your database.

```bash
# With make (recommended)
make migrateup

# Or manually
migrate -database "$DB_DSN" -path migrations up
```

### 4. Seed the Database (Optional but Recommended)

This command will create a default admin user and populate the database with product data.

```bash
# With make
make seed

# Or manually
go run ./cmd/seed/
```

### 5. Run the Application

```bash
# With make
make run

# Or manually
go run ./cmd/api/main.go
```

The server will start on `http://localhost:8080`.

## API Endpoints

### Authentication

| Method | Endpoint         | Description                                     | Access  |
| :----- | :--------------- | :---------------------------------------------- | :------ |
| `POST` | `/api/v1/auth/login` | Authenticate a user and get tokens.             | Public  |
| `POST` | `/api/v1/auth/refresh` | Get a new access token using a refresh token.   | Public  |
| `POST` | `/api/v1/auth/logout`  | Log out by revoking the current refresh token.  | User    |

### User

| Method   | Endpoint                  | Description                                      | Access |
| :------- | :------------------------ | :----------------------------------------------- | :----- |
| `POST`   | `/api/v1/users`           | Register a new user.                             | Public |
| `GET`    | `/api/v1/users/profile`   | Get the authenticated user's profile.            | User   |
| `PUT`    | `/api/v1/users/profile`   | Update the authenticated user's profile.         | User   |
| `PUT`    | `/api/v1/users/password`  | Change the authenticated user's password.        | User   |
| `DELETE` | `/api/v1/users/account`   | Delete the authenticated user's account.         | User   |
| `GET`    | `/api/v1/auth/sessions`   | View all active sessions for the user.           | User   |
| `DELETE` | `/api/v1/auth/sessions`   | Log out from all devices (revoke all sessions).  | User   |
| `DELETE` | `/api/v1/auth/sessions/{sessionId}` | Log out from a specific session.                | User   |

### Address Management

| Method   | Endpoint                          | Description                                      | Access |
| :------- | :-------------------------------- | :----------------------------------------------- | :----- |
| `POST`   | `/api/v1/addresses`              | Create a new address.                            | User   |
| `GET`    | `/api/v1/addresses`              | List all addresses for the user.                 | User   |
| `GET`    | `/api/v1/addresses/{id}`         | Get details for a specific address.              | User   |
| `PUT`    | `/api/v1/addresses/{id}`         | Update an address.                               | User   |
| `DELETE` | `/api/v1/addresses/{id}`         | Delete an address.                               | User   |
| `PUT`    | `/api/v1/addresses/{id}/set-default` | Set an address as default shipping/billing.    | User   |

### Products & Categories

| Method | Endpoint                    | Description                          | Access |
| :----- | :-------------------------- | :----------------------------------- | :----- |
| `GET`  | `/api/v1/products`          | Get a list of all products.          | Public |
| `GET`  | `/api/v1/products/{productId}` | Get details for a single product.    | Public |
| `GET`  | `/api/v1/categories`        | Get a list of all product categories.| Public |

### Cart

| Method   | Endpoint                       | Description                               | Access                |
| :------- | :----------------------------- | :---------------------------------------- | :-------------------- |
| `GET`    | `/api/v1/cart`                 | Get the current user's cart.              | User or Anonymous     |
| `POST`   | `/api/v1/cart/items`           | Add an item to the cart.                  | User or Anonymous     |
| `PATCH`  | `/api/v1/cart/items/{productId}` | Update an item's quantity in the cart.    | User or Anonymous     |
| `DELETE` | `/api/v1/cart/items/{productId}` | Remove an item from the cart.             | User or Anonymous     |

### Admin

| Method   | Endpoint                     | Description                    | Access |
| :------- | :--------------------------- | :----------------------------- | :----- |
| `GET`    | `/api/v1/admin/users`        | Get a list of all users.       | Admin  |
| `POST`   | `/api/v1/admin/users`        | Create a new user.             | Admin  |
| `PUT`    | `/api/v1/admin/users/{userId}` | Update a user's details.       | Admin  |
| `DELETE` | `/api/v1/admin/users/{userId}` | Delete a user.                 | Admin  |

## Future Work & Roadmap

-   [ ] **Comprehensive Testing:** Implement unit tests for services and repositories, and integration tests for handlers.
-   [ ] **Containerization:** Add a `Dockerfile` and `docker-compose.yml` for easy setup and deployment.
-   [ ] **CI/CD Pipeline:** Set up GitHub Actions to automate testing and building.
-   [ ] **API Documentation:** Generate OpenAPI (Swagger) documentation for the API.
-   [ ] **Order Management:** Implement endpoints for creating and viewing orders.
-   [ ] **Payment Integration:** Integrate a payment gateway like Stripe.