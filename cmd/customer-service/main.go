package main

import (
	"context"
	"net"
	"os"

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
	shutdown := otel.Init("customer-service")
	defer shutdown(context.Background())

	db, _ := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	r := repo.New(db)
	svc := service.New(r)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	pb.RegisterCustomerServiceServer(grpcServer, cgrpc.New(svc))

	lis, _ := net.Listen("tcp", ":9090")
	grpcServer.Serve(lis)
}
