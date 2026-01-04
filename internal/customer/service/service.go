package service

import (
	"context"

	"transline.kz/internal/customer/repo"
)

type Service struct {
	repo *repo.Repo
}

func New(repo *repo.Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpsertCustomer(ctx context.Context, idn string) (*repo.Customer, error) {
	return s.repo.Upsert(ctx, idn)
}
