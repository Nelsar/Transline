# Transline - Shipment Management API

Microservices platform for shipment management with gRPC communication, PostgreSQL, OpenTelemetry tracing.
## Database Migrations

This repository includes SQL migration files in the `migrations/` directory. Use the Makefile targets which leverage the `migrate/migrate` Docker image to apply migrations.
Examples:

```bash
make migrate-up
make migrate-down
```
If you prefer the CLI tool, install `golang-migrate` and run:

```bash
migrate -path ./migrations -database "$DATABASE_URL" up
```
## Logging & Timeouts

- Services use structured logging via `slog` (set in `cmd/*/main.go`).
- gRPC calls from Shipment service have a 2s timeout; DB ping uses a 5s timeout.
## Development

Build services locally:

```bash
make build
```

See `Makefile` for the `migrate-*` helpers.
# Transline - Shipment Management API

Microservices platform for shipment management with gRPC communication, PostgreSQL, OpenTelemetry tracing.

## Quick Start

```bash
docker-compose up
```

Access:
- Shipment API: http://localhost:8080
- Jaeger UI: http://localhost:16686

## Create Shipment

```bash
curl -X POST http://localhost:8080/api/v1/shipments \
  -H "Content-Type: application/json" \
  -d '{"route":"ALMATY→ASTANA","price":120000,"customer":{"idn":"990101123456"}}'
