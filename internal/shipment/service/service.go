package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

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

	// Upsert customer через gRPC (с таймаутом)
	grpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cus, err := s.customerGRPC.UpsertCustomer(grpcCtx, in.IDN)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert customer: %w", err)
	}

	// Parse customer ID from string to UUID
	customerID, err := uuid.Parse(cus.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid customer id format: %w", err)
	}

	// Создание shipment
	sh, err := s.repo.Create(
		ctx,
		customerID,
		in.Route,
		in.Price,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Возврат результата
	return &CreateShipmentResult{
		ID:         sh.ID,
		Status:     sh.Status,
		CustomerID: sh.CustomerID,
	}, nil
}
