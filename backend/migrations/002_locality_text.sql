-- Replace localities table with free-text locality on clients and service_orders.
-- Safe for fresh installs (001 already has locality TEXT) and existing DBs.

ALTER TABLE clients ADD COLUMN IF NOT EXISTS locality TEXT;
ALTER TABLE service_orders ADD COLUMN IF NOT EXISTS locality TEXT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'localities'
  ) THEN
    IF EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'clients' AND column_name = 'locality_id'
    ) THEN
      UPDATE clients c
      SET locality = l.name
      FROM localities l
      WHERE c.locality_id = l.id
        AND (c.locality IS NULL OR c.locality = '');
    END IF;

    IF EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'service_orders' AND column_name = 'locality_id'
    ) THEN
      UPDATE service_orders o
      SET locality = l.name
      FROM localities l
      WHERE o.locality_id = l.id
        AND (o.locality IS NULL OR o.locality = '');
    END IF;
  END IF;
END $$;

ALTER TABLE clients DROP COLUMN IF EXISTS locality_id;
ALTER TABLE service_orders DROP COLUMN IF EXISTS locality_id;
DROP TABLE IF EXISTS localities;
