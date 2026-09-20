-- Phase 2: WhatsApp templates and tariffs.

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
