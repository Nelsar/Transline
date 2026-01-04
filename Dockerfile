# =========================
# Build stage
# =========================
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# git нужен для go mod
RUN apk add --no-cache git

# go.mod / go.sum отдельно — для кеша
COPY go.mod go.sum ./
RUN go mod download

# копируем весь проект
COPY . .

# сборка shipment-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o shipment-service ./cmd/shipment-service

# сборка customer-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o customer-service ./cmd/customer-service

# =========================
# Runtime stage
# =========================
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# копируем бинарники
COPY --from=builder /app/shipment-service /app/shipment-service
COPY --from=builder /app/customer-service /app/customer-service

# контейнер стартует командой из docker-compose
