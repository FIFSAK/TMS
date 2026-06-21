# Shipment Tracking Microservice (TMS)

A REST/HTTP microservice for managing shipments and tracking status changes during transportation.

## Architecture

The project follows **Clean Architecture** principles with clear separation of concerns:

```
internal/
├── domain/shipment/     # Pure domain logic (entities, value objects, repository interface)
│   ├── entity.go        # Shipment and Event domain models
│   ├── status.go        # Status type with transition rules
│   └── repository.go    # Repository interface (port)
├── service/shipment/    # Application/use-case layer (business orchestration)
├── repository/sqlite/   # Infrastructure: SQLite implementation (adapter)
├── handler/rest/        # Transport: REST/JSON handler (adapter)
├── app/                 # Application bootstrap and lifecycle
└── config/              # Configuration management
pkg/
├── server/              # HTTP server lifecycle management
├── store/               # Database connection and migrations
└── log/                 # Structured logging (zap)
migrations/
└── sqlite/              # SQL migration files
```

**Key design decisions:**
- Domain layer has zero framework dependencies — testable in isolation
- Repository interface defined in the domain package (Dependency Inversion)
- Status transitions enforced at the domain level via a state machine
- REST handler converts between JSON DTOs and domain types, keeping transport concerns separate
- Routing uses [`go-chi/chi`](https://github.com/go-chi/chi) — a lightweight, idiomatic, `net/http`-compatible router (handlers stay plain `http.HandlerFunc`). Routes are mounted under a versioned `/api/v1` group with a middleware stack: request ID, real IP, zap request logging, and panic recovery

## Shipment Lifecycle

Shipments follow a strict status state machine:

```
pending → picked_up → in_transit → delivered
   ↓          ↓            ↓
cancelled  cancelled    cancelled
```

- `delivered` and `cancelled` are terminal states (no further transitions)
- All status changes are recorded as events, providing a full audit trail
- The current shipment status always reflects the latest valid event

## Prerequisites

- Go 1.24+

No external database required — SQLite is embedded and runs locally (file-based).

## Running the Service

1. Copy and configure environment:
   ```bash
   cp .env.dist .env
   # Optionally edit .env to change the HTTP port or SQLite database file path
   ```

2. Run the service:
   ```bash
   go run ./cmd/tms
   # or: make run
   ```

   The HTTP server starts on `:8080` by default (configurable via `HTTP_PORT`).

3. Exercise the API with the bundled demo client:
   ```bash
   go run ./cmd/client
   # or: make client
   ```

## Configuration

| Variable       | Description                  | Default                          |
|----------------|------------------------------|----------------------------------|
| `APP_MODE`     | Application mode             | `dev`                            |
| `HTTP_PORT`    | HTTP server listen address   | `:8080`                          |
| `SQLITE_DSN`   | SQLite database file path    | `tms.db`                         |

## REST API

All shipment endpoints are mounted under the versioned `/api/v1` group.
`GET /healthz` lives at the root (outside versioning).

| Method & Path                         | Description                               |
|---------------------------------------|-------------------------------------------|
| `POST /api/v1/shipments`              | Create a new shipment (starts as pending) |
| `GET /api/v1/shipments`               | List all shipments                        |
| `GET /api/v1/shipments/{id}`          | Retrieve a shipment by ID                 |
| `POST /api/v1/shipments/{id}/events`  | Add a status change event                 |
| `GET /api/v1/shipments/{id}/events`   | Get event history for a shipment          |
| `GET /healthz`                        | Health check                              |

All request and response bodies are JSON. Status values are strings: `pending`,
`picked_up`, `in_transit`, `delivered`, `cancelled`.

### Examples

```bash
# Create a shipment
curl -X POST http://localhost:8080/api/v1/shipments \
  -H 'Content-Type: application/json' \
  -d '{
    "reference_number": "SHP-2026-001",
    "origin": "New York, NY",
    "destination": "Los Angeles, CA",
    "driver_name": "John Smith",
    "unit_number": "TRUCK-42",
    "shipment_amount": 5000.00,
    "driver_revenue": 1500.00
  }'

# Get a shipment
curl http://localhost:8080/api/v1/shipments/<id>

# Add a status event
curl -X POST http://localhost:8080/api/v1/shipments/<id>/events \
  -H 'Content-Type: application/json' \
  -d '{"status": "picked_up", "comment": "Driver picked up the shipment"}'

# Get event history
curl http://localhost:8080/api/v1/shipments/<id>/events

# List all shipments
curl http://localhost:8080/api/v1/shipments
```

Error responses use `{"error": "..."}` with an appropriate HTTP status code
(`400` validation/invalid transition, `404` not found, `500` internal).

## Assumptions

- UUIDs are used for shipment and event IDs (generated server-side)
- Shipment amounts and driver revenue are stored as floating-point values
- The service does not implement authentication/authorization (out of scope)
- Migrations run automatically on startup
- Events are append-only; the latest event determines current status
