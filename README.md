# Task Scheduler API

A production-ready, containerized REST API for managing tasks. Built in Go using the Gin framework, it leverages PostgreSQL for persistent storage, Redis for high-performance cache-aside retrieval, and is fully orchestrated using Docker Compose.

---

## Features

- **Robust REST API**: Endpoints for creating, retrieving, and updating tasks.
- **Cache-Aside Pattern**: Automatically caches task details in Redis upon retrieval to reduce database load. Invalidates cache on update to guarantee consistency.
- **Race Condition Prevention**: Employs row-level locking (`SELECT ... FOR UPDATE` transactions) during updates.
- **Secure Configuration**: Externalizes credentials using `.env` environment files (fully ignored by Git).
- **Production-Grade Containers**: Optimized, secure multi-stage Alpine Docker builds running as a non-root user.
- **Self-Healing Infrastructure**: Sequences service initialization in Docker Compose using healthchecks so the app only boots when Postgres and Redis are fully online.

---

## Tech Stack

- **Language**: Go (Golang 1.25)
- **Web Framework**: Gin Gonic
- **Primary Database**: PostgreSQL 16
- **Cache Store**: Redis 7
- **Orchestration**: Docker Compose
- **Container Base**: Alpine Linux

---

## Getting Started

### Prerequisites
Make sure you have [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed on your machine.

### Installation & Run

1. **Clone the Repository**:
   ```bash
   git clone <your-repository-url>
   cd TaskScheduler
   ```

2. **Set Up Environments**:
   Copy the example environment template to create your local variables:
   ```bash
   cp .env.example .env
   ```
   *(Optional)* Edit `.env` to configure your custom credentials.

3. **Spin Up the Containers**:
   ```bash
   docker compose up --build
   ```
   This command automatically builds the Go application, pulls Postgres/Redis images, validates healthchecks, runs database table migrations, and starts the API on port `8080`.

---

## API Endpoints

### 1. Health Check
* **Route**: `GET /health`
* **Description**: Verifies the Go application's operational status.
* **Response (200 OK)**:
  ```json
  {"status": "healthy"}
  ```

### 2. Create Task
* **Route**: `POST /tasks`
* **Description**: Creates a new task. The `status` defaults to `pending` (overrides any client-provided status), and `id` and `created_at` are server-generated.
* **Payload**:
  ```json
  {
    "title": "Buy groceries",
    "description": "Milk, eggs, and bread"
  }
  ```
* **Response (201 Created)**:
  ```json
  {
    "id": "20260627134522",
    "title": "Buy groceries",
    "description": "Milk, eggs, and bread",
    "status": "pending",
    "created_at": "2026-06-27T13:45:22.123Z"
  }
  ```

### 3. Get Task (Cached)
* **Route**: `GET /tasks/:id`
* **Description**: Retrieves details for a specific task. Employs Redis cache lookup (hits serve instantly; misses fall back to Postgres and write back to cache).
* **Response (200 OK)**:
  ```json
  {
    "id": "20260627134522",
    "title": "Buy groceries",
    "description": "Milk, eggs, and bread",
    "status": "pending",
    "created_at": "2026-06-27T13:45:22.123Z"
  }
  ```

### 4. Update Task
* **Route**: `PUT /tasks/:id`
* **Description**: Updates task title, description, or status (e.g. `in-progress`, `completed`). Invalidates the Redis cache.
* **Payload**:
  ```json
  {
    "title": "Buy groceries (Urgent)",
    "status": "in-progress"
  }
  ```
* **Response (200 OK)**:
  ```json
  {
    "id": "20260627134522",
    "title": "Buy groceries (Urgent)",
    "description": "Milk, eggs, and bread",
    "status": "in-progress",
    "created_at": "2026-06-27T13:45:22.123Z"
  }
  ```
