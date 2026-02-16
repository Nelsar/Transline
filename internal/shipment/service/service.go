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

	if len(in.Route) > 255 {
		return nil, errors.New("route is too long (max 255 chars)")
	}

	if in.Price <= 0 {
		return nil, errors.New("price must be positive")
	}

	if in.Price > 1e10 {
		return nil, errors.New("price is too high")
	}

	if !regexp.MustCompile(`^\d{12}$`).MatchString(in.IDN) {
		return nil, errors.New("invalid idn format (must be 12 digits)")
	}

	// Upsert customer через gRPC
	cus, err := s.customerGRPC.UpsertCustomer(ctx, in.IDN)
	if err != nil {
		return nil, errors.New("failed to upsert customer: " + err.Error())
	}

	// Parse customer ID from string to UUID
	customerID, err := uuid.Parse(cus.Id)
	if err != nil {
		return nil, errors.New("invalid customer id format: " + err.Error())
	}

	// Создание shipment
	sh, err := s.repo.Create(
		ctx,
		customerID,
		in.Route,
		in.Price,
	)
	if err != nil {
		return nil, errors.New("failed to create shipment: " + err.Error())
	}

	// Возврат результата
	return &CreateShipmentResult{
		ID:         sh.ID,
		Status:     sh.Status,
		CustomerID: sh.CustomerID,
	}, nil
}
