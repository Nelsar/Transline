package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "transline.kz/api/proto/customerpb"
	"transline.kz/internal/otel"
	shgrpc "transline.kz/internal/shipment/grpc"
	shhttp "transline.kz/internal/shipment/http"
	"transline.kz/internal/shipment/repo"
	shservice "transline.kz/internal/shipment/service"
)

func main() {
	// OpenTelemetry
	shutdown := otel.Init("shipment-service")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("error shutting down otel: %v", err)
		}
	}()

	// PostgreSQL
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer db.Close()

	// Check DB connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.Ping(ctx); err != nil {
		cancel()
		log.Fatalf("db ping error: %v", err)
	}
	cancel()

	// gRPC client (через Envoy, с OTel StatsHandler)
	conn, err := grpc.Dial(
		"envoy:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("grpc dial error: %v", err)
	}
	defer conn.Close()

	customerClient := shgrpc.New(
		pb.NewCustomerServiceClient(conn),
	)

	// Application layers
	repository := repo.New(db)
	service := shservice.New(repository, customerClient)
	handler := shhttp.New(service)

	// HTTP router
	mux := http.NewServeMux()
	mux.Handle(
		"/api/v1/shipments",
		otelhttp.NewHandler(
			http.HandlerFunc(handler.Create),
			"CreateShipment",
		),
	)

	// HTTP server with graceful shutdown
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("http server shutdown error: %v", err)
		}
	}()

	log.Println("shipment-service listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server error: %v", err)
	}
	log.Println("shipment-service stopped")
}
