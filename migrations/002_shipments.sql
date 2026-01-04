-- 002_shipments.sql
CREATE TABLE shipments (
  id UUID PRIMARY KEY,
  route TEXT NOT NULL,
  price NUMERIC NOT NULL,
  status TEXT NOT NULL DEFAULT 'CREATED',
  customer_id UUID NOT NULL REFERENCES customers(id),
  created_at TIMESTAMP DEFAULT now()
);
