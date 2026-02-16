package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	pb "transline.kz/api/proto/customerpb"
	cgrpc "transline.kz/internal/customer/grpc"
	"transline.kz/internal/customer/repo"
	"transline.kz/internal/customer/service"
	"transline.kz/internal/otel"
)

func main() {
	// OpenTelemetry
	shutdown := otel.Init("customer-service")
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

	r := repo.New(db)
	svc := service.New(r)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	pb.RegisterCustomerServiceServer(grpcServer, cgrpc.New(svc))

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("listener error: %v", err)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("shutdown signal received")
		grpcServer.GracefulStop()
	}()

	log.Println("customer-service listening on :9090")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc server error: %v", err)
	}
}
