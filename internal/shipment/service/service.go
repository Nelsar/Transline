package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
	shgrpc "transline.kz/internal/shipment/grpc"
	"transline.kz/internal/shipment/repo"
)

type Service struct {
	repo         *repo.Repo
	customerGRPC *shgrpc.Client
}

func New(
	repo *repo.Repo,
	customerGRPC *shgrpc.Client,
) *Service {
	return &Service{
		repo:         repo,
		customerGRPC: customerGRPC,
	}
}

type CreateShipmentInput struct {
	Route string
	Price float64
	IDN   string
}

type CreateShipmentResult struct {
	ID         uuid.UUID
	Status     string
	CustomerID uuid.UUID
}

func (s *Service) CreateShipment(
	ctx context.Context,
	in CreateShipmentInput,
) (*CreateShipmentResult, error) {

	// Бизнес-валидация
	if in.Route == "" {
		return nil, errors.New("route is required")
	}

	if in.Price <= 0 {
		return nil, errors.New("price must be positive")
	}

	if !regexp.MustCompile(`^\d{12}$`).MatchString(in.IDN) {
		return nil, errors.New("invalid idn")
	}

	//Upsert customer через gRPC
	cus, err := s.customerGRPC.UpsertCustomer(ctx, in.IDN)
	if err != nil {
		return nil, err
	}

	//Создание shipment
	sh, err := s.repo.Create(
		ctx,
		cus.Id,
		in.Route,
		in.Price,
	)
	if err != nil {
		return nil, err
	}

	//Возврат результата
	return &CreateShipmentResult{
		ID:         sh.ID,
		Status:     sh.Status,
		CustomerID: sh.CustomerID,
	}, nil
}
