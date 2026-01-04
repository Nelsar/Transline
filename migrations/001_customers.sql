-- 001_customers.sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE customers (
  id UUID PRIMARY KEY,
  idn TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT now()
);
