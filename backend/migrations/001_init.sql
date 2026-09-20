-- Phase 1 schema (Postgres). Includes columns that lived in Node migrate.js.

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  full_name TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('admin', 'coordinador', 'tecnico')),
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS technicians (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT UNIQUE REFERENCES users(id),
  name TEXT NOT NULL,
  phone TEXT,
  specialty TEXT,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS providers (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('proveedor', 'prestador')),
  contact TEXT,
  notes TEXT
);

CREATE TABLE IF NOT EXISTS clients (
  id BIGSERIAL PRIMARY KEY,
  provider_id BIGINT REFERENCES providers(id),
  external_id TEXT,
  name TEXT NOT NULL,
  phone TEXT,
  email TEXT,
  locality TEXT,
  address TEXT,
  notes TEXT,
  source TEXT NOT NULL DEFAULT 'manual',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS product_types (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS products (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  sku TEXT,
  product_type_id BIGINT REFERENCES product_types(id),
  provider_id BIGINT REFERENCES providers(id),
  stock INTEGER NOT NULL DEFAULT 0,
  price DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS service_statuses (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  sort_order INTEGER NOT NULL,
  color TEXT NOT NULL,
  is_closed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS service_orders (
  id BIGSERIAL PRIMARY KEY,
  client_id BIGINT NOT NULL REFERENCES clients(id),
  technician_id BIGINT REFERENCES technicians(id),
  product_type_id BIGINT REFERENCES product_types(id),
  product_id BIGINT REFERENCES products(id),
  product_label TEXT,
  locality TEXT,
  provider_id BIGINT REFERENCES providers(id),
  status_id BIGINT NOT NULL REFERENCES service_statuses(id),
  title TEXT NOT NULL,
  description TEXT,
  scheduled_date DATE,
  scheduled_time TEXT,
  hours DOUBLE PRECISION NOT NULL DEFAULT 0,
  km DOUBLE PRECISION NOT NULL DEFAULT 0,
  parts_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
  parts_sale DOUBLE PRECISION NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS followup_rules (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  trigger_type TEXT NOT NULL CHECK (trigger_type IN ('status', 'elapsed_hours')),
  status_id BIGINT REFERENCES service_statuses(id),
  hours INTEGER,
  assign_role TEXT CHECK (assign_role IN ('admin', 'coordinador', 'tecnico')),
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS followups (
  id BIGSERIAL PRIMARY KEY,
  service_order_id BIGINT NOT NULL REFERENCES service_orders(id),
  rule_id BIGINT REFERENCES followup_rules(id),
  assigned_user_id BIGINT REFERENCES users(id),
  due_at TIMESTAMPTZ,
  notes TEXT,
  status TEXT NOT NULL DEFAULT 'pendiente' CHECK (status IN ('pendiente', 'hecho')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS whatsapp_templates (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  body TEXT NOT NULL,
  status_id BIGINT REFERENCES service_statuses(id),
  auto_open BOOLEAN NOT NULL DEFAULT FALSE,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS tariffs (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  scope TEXT NOT NULL CHECK (scope IN ('proveedor', 'tecnico', 'general')),
  provider_id BIGINT REFERENCES providers(id),
  technician_id BIGINT REFERENCES technicians(id),
  income_fixed DOUBLE PRECISION NOT NULL DEFAULT 0,
  income_per_hour DOUBLE PRECISION NOT NULL DEFAULT 0,
  income_per_km DOUBLE PRECISION NOT NULL DEFAULT 0,
  income_parts_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
  cost_fixed DOUBLE PRECISION NOT NULL DEFAULT 0,
  cost_per_hour DOUBLE PRECISION NOT NULL DEFAULT 0,
  cost_per_km DOUBLE PRECISION NOT NULL DEFAULT 0,
  cost_parts_pct DOUBLE PRECISION NOT NULL DEFAULT 100,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_orders_date ON service_orders(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_orders_tech ON service_orders(technician_id);
CREATE INDEX IF NOT EXISTS idx_clients_provider ON clients(provider_id);
CREATE INDEX IF NOT EXISTS idx_followups_status ON followups(status);
