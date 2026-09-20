-- Normalize spreadsheet workflow: clients + service_requests, unit_types, technician codes.

ALTER TABLE clients ADD COLUMN IF NOT EXISTS phone_alt TEXT;

ALTER TABLE technicians ADD COLUMN IF NOT EXISTS code TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS uq_technicians_code
  ON technicians (code)
  WHERE code IS NOT NULL AND btrim(code) <> '';

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'product_types'
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'unit_types'
  ) THEN
    ALTER TABLE product_types RENAME TO unit_types;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'service_orders'
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'service_requests'
  ) THEN
    ALTER TABLE service_orders RENAME TO service_requests;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service_requests' AND column_name = 'product_type_id'
  ) THEN
    ALTER TABLE service_requests RENAME COLUMN product_type_id TO unit_type_id;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service_requests' AND column_name = 'scheduled_date'
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service_requests' AND column_name = 'visit_date'
  ) THEN
    ALTER TABLE service_requests RENAME COLUMN scheduled_date TO visit_date;
  END IF;
END $$;

ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS received_at DATE;
ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS provider_order_ref TEXT;
ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS internal_order_no TEXT;
ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS request_kind TEXT;
ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS appliance_model TEXT;
ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS reported_failure TEXT;
ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS ops_notes TEXT;
ALTER TABLE service_requests ADD COLUMN IF NOT EXISTS diagnosis_notes TEXT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'service_requests_request_kind_check'
  ) THEN
    ALTER TABLE service_requests
      ADD CONSTRAINT service_requests_request_kind_check
      CHECK (request_kind IS NULL OR request_kind IN ('G', 'FG'));
  END IF;
END $$;

UPDATE service_requests
SET received_at = created_at::date
WHERE received_at IS NULL;

UPDATE service_requests
SET appliance_model = COALESCE(NULLIF(btrim(product_label), ''), NULLIF(btrim(title), ''))
WHERE appliance_model IS NULL;

UPDATE service_requests
SET reported_failure = description
WHERE reported_failure IS NULL AND description IS NOT NULL;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'followups' AND column_name = 'service_order_id'
  ) THEN
    ALTER TABLE followups RENAME COLUMN service_order_id TO service_request_id;
  END IF;
END $$;

DROP INDEX IF EXISTS idx_orders_date;
DROP INDEX IF EXISTS idx_orders_tech;
CREATE INDEX IF NOT EXISTS idx_requests_visit_date ON service_requests (visit_date);
CREATE INDEX IF NOT EXISTS idx_requests_tech ON service_requests (technician_id);
CREATE INDEX IF NOT EXISTS idx_requests_provider_order ON service_requests (provider_order_ref);
CREATE INDEX IF NOT EXISTS idx_requests_received ON service_requests (received_at);
CREATE INDEX IF NOT EXISTS idx_followups_request ON followups (service_request_id);

INSERT INTO service_statuses (name, sort_order, color, is_closed)
SELECT v.name, v.sort_order, v.color, v.is_closed
FROM (VALUES
  ('Coordinado', 8, '#0ea5e9', FALSE),
  ('En taller', 9, '#a855f7', FALSE),
  ('Anulado', 10, '#dc2626', TRUE),
  ('Cerrado', 11, '#334155', TRUE)
) AS v(name, sort_order, color, is_closed)
WHERE NOT EXISTS (
  SELECT 1 FROM service_statuses s WHERE s.name = v.name
);

INSERT INTO unit_types (name)
SELECT v.name
FROM (VALUES
  ('HELADERA'),
  ('LAVARROPAS'),
  ('LAVASECARROPAS'),
  ('SECARROPAS'),
  ('LAVAVAJILLAS'),
  ('MICROONDAS'),
  ('ANAFE'),
  ('HORNO ELECTRICO'),
  ('HORNO GAS'),
  ('COCINA'),
  ('PURIFICADOR'),
  ('FREIDORA POR AIRE')
) AS v(name)
WHERE NOT EXISTS (
  SELECT 1 FROM unit_types u WHERE upper(u.name) = v.name
);
