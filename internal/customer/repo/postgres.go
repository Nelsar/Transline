package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Customer struct {
	ID        string
	IDN       string
	CreatedAt time.Time
}

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Upsert(ctx context.Context, idn string) (*Customer, error) {
	row := r.db.QueryRow(ctx, `
    INSERT INTO customers (id, idn)
    VALUES (gen_random_uuid(), $1)
    ON CONFLICT (idn)
      DO UPDATE SET idn = EXCLUDED.idn
    RETURNING id, idn, created_at
  `, idn)

	c := Customer{}
	err := row.Scan(&c.ID, &c.IDN, &c.CreatedAt)
	return &c, err
}
