# Shipment Tracking Microservice (TMS)

A gRPC microservice for managing shipments and tracking status changes during transportation.

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
├── handler/grpc/        # Transport: gRPC handler (adapter)
├── app/                 # Application bootstrap and lifecycle
└── config/              # Configuration management
pkg/
├── pb/                  # Generated protobuf/gRPC code
├── server/              # gRPC server lifecycle management
├── store/               # Database connection and migrations
└── log/                 # Structured logging (zap)
proto/
└── shipment/v1/         # Protocol Buffer definitions
migrations/
└── sqlite/              # SQL migration files
```

**Key design decisions:**
- Domain layer has zero framework dependencies — testable in isolation
- Repository interface defined in the domain package (Dependency Inversion)
- Status transitions enforced at the domain level via a state machine
- gRPC handler converts between protobuf and domain types, keeping transport concerns separate

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
- protoc (for regenerating protobuf code)

No external database required — SQLite is embedded and runs locally (file-based).

## Running the Service

1. Copy and configure environment:
   ```bash
   cp .env.dist .env
   # Optionally edit .env to change the SQLite database file path
   ```

2. Run the service:
   ```bash
   go run cmd/tns/main.go
   ```

   The gRPC server starts on `:50051` by default (configurable via `GRPC_PORT`).

## Running Tests

```bash
# Domain and service tests (no database required)
go test ./internal/domain/shipment/ ./internal/service/shipment/ -v

# Or use make
make test
```

Tests cover:
- Shipment creation with validation
- All valid status transitions
- Invalid status transition rejection
- Terminal state enforcement
- Service-level business logic with mocked repository

## Configuration

| Variable       | Description                  | Default                          |
|----------------|------------------------------|----------------------------------|
| `APP_MODE`     | Application mode             | `dev`                            |
| `GRPC_PORT`    | gRPC server listen address   | `:50051`                         |
| `SQLITE_DSN`   | SQLite database file path    | `tms.db`                         |

## gRPC API

Defined in `proto/shipment/v1/shipment.proto`:

| RPC                | Description                          |
|--------------------|--------------------------------------|
| `CreateShipment`   | Create a new shipment (starts as pending) |
| `GetShipment`      | Retrieve shipment by ID              |
| `ListShipments`    | List all shipments                   |
| `AddShipmentEvent` | Add a status change event            |
| `GetShipmentEvents`| Get event history for a shipment     |

## Regenerating Protobuf Code

```bash
make proto
```

Requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc`.

## Assumptions

- UUIDs are used for shipment and event IDs (generated server-side)
- Shipment amounts and driver revenue are stored as floating-point values
- The service does not implement authentication/authorization (out of scope)
- Migrations run automatically on startup
- Events are append-only; the latest event determines current status
