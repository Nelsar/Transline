package grpc

import (
	"context"
	"time"

	pb "transline.kz/api/proto/customerpb"
	"transline.kz/internal/customer/service"
)

type Server struct {
	pb.UnimplementedCustomerServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) UpsertCustomer(ctx context.Context, req *pb.UpsertCustomerRequest) (*pb.CustomerResponse, error) {
	c, err := s.svc.UpsertCustomer(ctx, req.Idn)
	if err != nil {
		return nil, err
	}

	return &pb.CustomerResponse{
		Id:        c.ID,
		Idn:       c.IDN,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}, nil
}
