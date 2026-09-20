PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  full_name TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('admin', 'coordinador', 'tecnico')),
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS technicians (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER UNIQUE REFERENCES users(id),
  name TEXT NOT NULL,
  phone TEXT,
  specialty TEXT,
  active INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('proveedor', 'prestador')),
  contact TEXT,
  notes TEXT
);

CREATE TABLE IF NOT EXISTS clients (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider_id INTEGER REFERENCES providers(id),
  external_id TEXT,
  name TEXT NOT NULL,
  phone TEXT,
  email TEXT,
  locality TEXT,
  address TEXT,
  notes TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS product_types (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS products (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  sku TEXT,
  product_type_id INTEGER REFERENCES product_types(id),
  provider_id INTEGER REFERENCES providers(id),
  stock INTEGER NOT NULL DEFAULT 0,
  price REAL NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS service_statuses (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  sort_order INTEGER NOT NULL,
  color TEXT NOT NULL,
  is_closed INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS service_orders (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  client_id INTEGER NOT NULL REFERENCES clients(id),
  technician_id INTEGER REFERENCES technicians(id),
  product_type_id INTEGER REFERENCES product_types(id),
  product_id INTEGER REFERENCES products(id),
  locality TEXT,
  provider_id INTEGER REFERENCES providers(id),
  status_id INTEGER NOT NULL REFERENCES service_statuses(id),
  title TEXT NOT NULL,
  description TEXT,
  scheduled_date TEXT,
  scheduled_time TEXT,
  started_at TEXT,
  completed_at TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS followup_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  trigger_type TEXT NOT NULL CHECK (trigger_type IN ('status', 'elapsed_hours')),
  status_id INTEGER REFERENCES service_statuses(id),
  hours INTEGER,
  assign_role TEXT CHECK (assign_role IN ('admin', 'coordinador', 'tecnico')),
  active INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS followups (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  service_order_id INTEGER NOT NULL REFERENCES service_orders(id),
  rule_id INTEGER REFERENCES followup_rules(id),
  assigned_user_id INTEGER REFERENCES users(id),
  due_at TEXT,
  notes TEXT,
  status TEXT NOT NULL DEFAULT 'pendiente' CHECK (status IN ('pendiente', 'hecho')),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  completed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_orders_date ON service_orders(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_orders_tech ON service_orders(technician_id);
CREATE INDEX IF NOT EXISTS idx_clients_provider ON clients(provider_id);
CREATE INDEX IF NOT EXISTS idx_followups_status ON followups(status);
