# GoKart E-Commerce API: A Cloud-Native Microservices Backend

![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Architecture](https://img.shields.io/badge/architecture-event--driven-9cf)
![Deployment](https://img.shields.io/badge/deploy-Google%20Cloud%20Run-4285F4)
![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1.svg)

**GoKart** is a production-ready, cloud-native backend for a modern e-commerce platform, architected using Go. This project serves as a comprehensive demonstration of building a scalable, resilient, and secure system following professional software engineering principles.

The architecture intentionally evolves from a clean, layered monolith for core API services to a decoupled, event-driven model using serverless workers for asynchronous processing. This showcases a pragmatic approach to system design, applying complexity where it adds tangible value in resilience and scalability.

---

## Architectural Pillars & Key Features

This project was built with a strong emphasis on professional design patterns that address real-world engineering challenges.

### 1. Decoupled, Event-Driven Fulfillment
The core of the application is its asynchronous, message-driven order fulfillment pipeline. This architecture provides exceptional resilience and scalability.

-   **Serverless Workers:** Long-running processes (warehouse, shipping, delivery, notifications) are handled by a separate, internal-only microservice (`mega-worker`) deployed on Cloud Run.
-   **Asynchronous Orchestration:** **Google Cloud Tasks** is used as a durable, at-least-once task queue to orchestrate the multi-step fulfillment workflow, ensuring no orders are lost.
-   **Resilience by Design:** A failure in a downstream service (e.g., the notification worker's email provider is down) has **zero impact** on the critical path of order processing. The failed task is automatically retried without affecting other parts of the system.
-   **Scalability:** The public-facing API and the internal worker are deployed as separate Cloud Run services, allowing them to be scaled independently based on their specific loads.

*Diagram: Asynchronous Order Fulfillment Workflow (to be added)*

### 2. Secure by Design
Security is a foundational principle, not an afterthought, implemented through modern best practices.

-   **Secure Session Management:** Stateful sessions are managed using JWT access tokens (short-lived, stored in memory) and refresh tokens (long-lived, stored in **secure, `HttpOnly` cookies**), providing robust protection against XSS-based token theft.
-   **Service-to-Service Authentication:** All internal communication between microservices (e.g., worker-to-API status updates) is secured using **Google-signed OIDC identity tokens**, eliminating the need for static API keys.
-   **Least Privilege Principle:** Each service runs with a dedicated IAM Service Account that has the absolute minimum permissions required for its function.
-   **Password Security:** Passwords are never stored. They are hashed using the industry-standard `bcrypt` algorithm.

<h4>Authentication Flow Diagrams</h4>
<div style="display: flex; justify-content: space-between;">
  <img src="architecture/login_and_token.png" alt="Login and Token Handling" style="width: 48%;">
  <img src="architecture/refresh_token_handling.png" alt="Refresh Token Handling" style="width: 48%;">
</div>

### 3. Professional Engineering Practices
The project adheres to patterns that promote maintainability, testability, and developer velocity.

-   **Clean, Layered Architecture:** Follows the classic **Handler-Service-Repository** pattern to ensure a strict separation of concerns.
-   **Atomic Transactions:** All critical, multi-step database operations (e.g., order creation with stock decrement, cart merging, cleanup jobs) are executed within **atomic transactions** to guarantee data consistency.
-   **Configuration Management:** All environment-specific settings are externalized into `.env` files and loaded into typed structs, separating code from configuration (a Twelve-Factor App principle).
-   **Automated CI/CD:** A comprehensive **GitHub Actions** workflow intelligently builds, tests, and deploys only the services whose code has changed, which is essential for a monorepo structure.
-   **Seamless User Experience:** Features like the **anonymous-to-user cart merge** on login/register demonstrate a focus on practical, user-centric design.

*Diagram: Anonymous-to-User Cart Merging (to be added)*

---

## Tech Stack & Tooling

| Category | Technology / Tool | Rationale |
| :--- | :--- | :--- |
| **Language** | Go 1.24+ | Excellent performance, static typing, and a first-class ecosystem for cloud-native development. |
| **Web Framework** | [Chi (v5)](https://github.com/go-chi/chi) | A lightweight, idiomatic, and composable router that encourages good API design. |
| **Database** | PostgreSQL | A robust, open-source relational database with powerful features like `JSONB` and transactional integrity. |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) | For managing database schema evolution in a version-controlled, repeatable manner. |
| **Deployment** | [Google Cloud Run](https://cloud.google.com/run) | A fully-managed serverless platform that provides auto-scaling (including scale-to-zero) and a secure runtime. |
| **Task Queue** | [Google Cloud Tasks](https://cloud.google.com/tasks) | A durable, serverless task queue for orchestrating asynchronous, event-driven workflows. |
| **Containerization** | [Docker](https://www.docker.com/) | For creating reproducible, lightweight application images based on secure `distroless` base images. |
| **CI/CD** | [GitHub Actions](https://github.com/features/actions) | For automating the build, push, and deployment pipeline for the multi-service monorepo. |
| **Authentication** | JWT & OIDC Tokens | Modern standards for user session management and secure service-to-service communication. |

---

## Getting Started

Follow these instructions to get the complete project, including the database and all services, running locally.

### Prerequisites

-   Go 1.24+
-   Docker & Docker Compose
-   [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
-   `gcloud` CLI (for authenticating to Google Cloud for task creation)

### 1. Clone the Repository

```bash
git clone https://github.com/purushothdl/gokart-ecommerce.git
cd gokart-ecommerce
```

### 2. Configure Local Environment

The project requires two environment files for its two services.

1.  Copy the example files:
    ```bash
    cp api.env.example api.env
    cp worker.env.example worker.env
    ```

2.  **Edit `api.env` and `worker.env`:**
    -   **For Easy Local Setup:** To use the included Dockerized PostgreSQL database, set the `DB_DSN` in both files to:
        ```
        DB_DSN=postgres://postgres:password@postgres:5432/gokart?sslmode=disable
        ```
    -   Fill in your other credentials for Stripe, Resend, and Google Cloud as detailed in the files.
    -   For local development, the `ECOMMERCE_API_URL` and `MEGA_WORKER_URL` can be set to `http://localhost:8080` and `http://localhost:8081` respectively, as Docker Compose makes them accessible.

### 3. Run the Application Stack

1.  **Start all services:**
    ```bash
    docker-compose up --build
    ```
    This single command will:
    -   Start a PostgreSQL database container.
    -   Build the Docker image for the `api` server.
    -   Build the Docker image for the `mega-worker`.
    -   Start all three containers and connect them on an internal network.

2.  **Run Database Migrations:**
    In a separate terminal, run the `make` command to set up the database schema.
    ```bash
    make migrateup
    ```

3.  **Seed the Database (Optional but Recommended):**
    This command creates a default admin user (`admin@example.com` / `supersecretpassword`) and populates the database with product data.
    ```bash
    make seed
    ```

The **API Server** is now available at `http://localhost:8080`.
The **Worker Server** is listening internally at `http://localhost:8081`.

---

## API Endpoints

A comprehensive list of API endpoints can be found in `internal/server/routes.go`. Key routes include:
-   **Authentication:** `/api/v1/auth/{login, refresh, logout}`
-   **Users:** `/api/v1/users/{profile, password}`
-   **Products:** `/api/v1/{products, categories}`
-   **Cart:** `/api/v1/cart/**`
-   **Orders:** `/api/v1/orders/**`
-   **Admin:** `/api/v1/admin/**`

## Future Work & Roadmap

-   [ ] Implement a comprehensive test suite (unit, integration, e2e).
-   [ ] Generate OpenAPI (Swagger) documentation for the API.
-   [ ] Integrate a frontend application built with React or Vue.
-   [ ] Enhance search functionality with a dedicated engine like Elasticsearch.