# DemonDoof - Backend (Learning Go)

A Go backend for a turn-based tactical grid game, designed for learning and experimentation. It exposes REST APIs (Fiber), a WebSocket gateway, persists data in PostgreSQL, and implements JWT authentication.

---

## Features

- Fast HTTP APIs with Fiber v2
- WebSocket endpoint for real-time messaging (requires Bearer JWT)
- JWT-based authentication (HS256)
- PostgreSQL storage via pgx pool
- Structured logging with slog + tint
- Database migrations using goose (initial schema + seed data)
- Bruno API collection for testing

---

## Tech Stack

- Go
- Fiber
- PostgreSQL
- pgx
- JWT (HS256)
- bcrypt
- slog + tint (logging)
- goose (migrations)
- Bruno (API testing)

---

## API Overview

### HTTP Endpoints

- `GET /health` — Returns status of server and database
- `POST /api/v1/auth/register` — User registration
- `POST /api/v1/auth/login` — User login, returns a JWT
- `GET /api/v1/auth/me` — User profile (requires Bearer JWT)

### WebSocket

- Endpoint: `/ws` (requires Bearer JWT)
- Example: Send `{"type":"ping"}` — receives `{"type":"pong", data: ...}`

---

## Getting Started

1. **Start PostgreSQL:**
   ```sh
   docker compose up -d postgres
   ```
2. **Run migrations:**
   ```sh
   make migrate-up
   ```
3. **Run the server:**
   ```sh
   go run ./cmd/server/main.go
   ```
   Or with automatic restart:
   ```sh
   nodemon --exec go run ./cmd/server/main.go --signal SIGTERM
   ```

---

## Project Structure

```
.
├── .env.dev                # Environment variables for development
├── Collection/             # Bruno API collections for endpoint testing
│   └── DemonDoof-Ultimate/ # Auth, health, and environment test cases
├── cmd/
│   └── server/             # Entry point for starting the server
├── internal/
│   ├── features/
│   │   └── users/          # User domain logic, repository, service
│   ├── server/             # Dependency injection and server setup
│   └── transport/
│       ├── http/           # HTTP controllers, routes, DTOs
│       └── ws/             # WebSocket routing and handlers
├── migrations/             # SQL migration scripts (schema, seed data)
├── pkg/
│   ├── auth/               # JWT handling, authentication helpers
│   ├── config/             # Configuration management
│   ├── db/                 # Database connection logic
│   ├── logger/             # Logging setup and helpers
│   └── middleware/         # HTTP middleware (e.g., authentication)
└── tests/
    ├── integration/        # Integration tests (todo)
    └── unit/               # Unit tests (todo)
```

₍^. .^₎⟆
