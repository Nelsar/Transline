package main

import (
	"context"
	"log/slog"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			slog.Error("error shutting down otel", "err", err)
		}
	}()

	// PostgreSQL
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("db connect error", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Check DB connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.Ping(ctx); err != nil {
		cancel()
		slog.Error("db ping error", "err", err)
		os.Exit(1)
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
		slog.Error("listener error", "err", err)
		os.Exit(1)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("shutdown signal received")
		grpcServer.GracefulStop()
	}()

	slog.Info("customer-service listening", "addr", ":9090")
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("grpc server error", "err", err)
		os.Exit(1)
	}
}
