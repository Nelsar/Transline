# Makefile for common tasks

.PHONY: migrate-up migrate-down migrate-force build

# Run migrations using golang-migrate docker image
migrate-up:
	@echo "Running migrations up..."
	docker run --rm -v $(PWD)/migrations:/migrations --network host migrate/migrate \
	  -path=/migrations -database "${DATABASE_URL}" up

migrate-down:
	@echo "Running migrations down..."
	docker run --rm -v $(PWD)/migrations:/migrations --network host migrate/migrate \
	  -path=/migrations -database "${DATABASE_URL}" down

migrate-force:
	@echo "Force migration to version"
	if [ -z "$(version)" ]; then echo "Usage: make migrate-force version=<num>"; exit 1; fi
	docker run --rm -v $(PWD)/migrations:/migrations --network host migrate/migrate \
	  -path=/migrations -database "${DATABASE_URL}" force $(version)

build:
	go build -o shipment-service ./cmd/shipment-service
	go build -o customer-service ./cmd/customer-service
