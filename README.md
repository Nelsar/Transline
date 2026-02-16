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
