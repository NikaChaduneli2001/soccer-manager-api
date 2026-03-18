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

## Migrations

Migrations run **automatically** when the API starts: after connecting to the database, the app runs all SQL files in `migrations/` in filename order (`001_...`, `002_...`, etc.). No separate step is required when using `make run` or Docker.

To run migrations **manually** (e.g. against a running PostgreSQL before starting the API):

```bash
# With Docker Compose (DB already running)
docker-compose up -d db
for f in migrations/*.sql; do psql -h localhost -p 5432 -U soccer -d soccer_manager -f "$f"; done

# Or with a single psql command (from project root)
psql -h localhost -p 5432 -U soccer -d soccer_manager -f migrations/001_create_users.sql
psql -h localhost -p 5432 -U soccer -d soccer_manager -f migrations/002_create_teams.sql
psql -h localhost -p 5432 -U soccer -d soccer_manager -f migrations/003_create_players.sql
psql -h localhost -p 5432 -U soccer -d soccer_manager -f migrations/004_create_transfer_listings.sql
```

Use the same host, port, user, and database as in your `.env` (or docker-compose). The password will be prompted unless you set `PGPASSWORD`.

## API documentation (Swagger)

With the API running, open **http://localhost:3000/swagger/index.html** for interactive Swagger UI. All endpoints are described there, including request/response schemas and the `Authorization: Bearer <token>` security scheme for protected routes.

To regenerate the Swagger spec from code annotations, install the CLI and run:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger-docs
```

## Environment

Copy `.env.example` to `.env` and adjust. Defaults work with `docker-compose` for DB connection.

## License

MIT
