# Soccer Manager API

RESTful API for a fantasy football manager application. Built with Go.

## Project structure

```
.
├── cmd/api/          # Application entrypoint
├── config/           # Configuration (env-based)
├── controller/       # HTTP request handling (handles requests, calls service)
├── handler/          # Routing (registers routes, delegates to controller)
├── middleware/       # HTTP middleware (auth, logging, etc.)
├── models/           # Domain models
├── pkg/              # Shared packages (response helpers, etc.)
├── repository/       # Data access layer
├── service/          # Business logic
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Prerequisites

- Go 1.22+
- Docker & Docker Compose (for full stack with PostgreSQL)

## Quick start

### Run locally (API only; no database)

```bash
make run
```

The server starts on `http://localhost:3000`. Health check: `GET /health`.

### Run with Docker (API + PostgreSQL)

```bash
make docker-up-build
```

API: `http://localhost:3000`, PostgreSQL: `localhost:5432` (user: `soccer`, password: `soccer`, db: `soccer_manager`).

### Makefile targets

| Target             | Description                    |
|--------------------|--------------------------------|
| `make run`         | Run the API locally            |
| `make build`       | Build binary to `bin/api`      |
| `make test`        | Run tests                      |
| `make tidy`        | Download and tidy dependencies |
| `make docker-up`   | Start containers               |
| `make docker-down` | Stop containers                |
| `make docker-up-build` | Build and start with Compose |
| `make clean`       | Remove `bin/`                  |

## Environment

Copy `.env.example` to `.env` and adjust. Defaults work with `docker-compose` for DB connection.

## License

MIT
