package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Shipment struct {
	ID         uuid.UUID
	Route      string
	Price      float64
	Status     string
	CustomerID uuid.UUID
	CreatedAt  time.Time
}

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, customerID uuid.UUID, route string, price float64) (*Shipment, error) {
	id := uuid.New()
	row := r.db.QueryRow(ctx, `
    INSERT INTO shipments (id, route, price, customer_id)
    VALUES ($1,$2,$3,$4)
    RETURNING id, route, price, status, customer_id, created_at
  `, id, route, price, customerID)

	s := Shipment{}
	err := row.Scan(&s.ID, &s.Route, &s.Price, &s.Status, &s.CustomerID, &s.CreatedAt)
	return &s, err
}
