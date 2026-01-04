package main

import (
	"context"
	"log"
	"net/http"
	"os"
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
	//OpenTelemetry
	shutdown := otel.Init("shipment-service")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	//PostgreSQL
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer db.Close()

	//gRPC client (через Envoy, с OTel StatsHandler)
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

	//Слои приложения
	repository := repo.New(db)
	service := shservice.New(repository, customerClient)
	handler := shhttp.New(service)

	//HTTP router
	mux := http.NewServeMux()
	mux.Handle(
		"/api/v1/shipments",
		otelhttp.NewHandler(
			http.HandlerFunc(handler.Create),
			"CreateShipment",
		),
	)

	// 6️⃣ HTTP server
	log.Println("shipment-service listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}
