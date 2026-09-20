import { db } from './db.js';

function columns(table) {
  return db.prepare(`PRAGMA table_info(${table})`).all().map((c) => c.name);
}

function addColumn(table, name, ddl) {
  if (!columns(table).includes(name)) {
    db.exec(`ALTER TABLE ${table} ADD COLUMN ${name} ${ddl}`);
  }
}

export function migrate() {
  addColumn('clients', 'source', "TEXT NOT NULL DEFAULT 'manual'");
  addColumn('service_orders', 'hours', 'REAL NOT NULL DEFAULT 0');
  addColumn('service_orders', 'km', 'REAL NOT NULL DEFAULT 0');
  addColumn('service_orders', 'parts_cost', 'REAL NOT NULL DEFAULT 0');
  addColumn('service_orders', 'parts_sale', 'REAL NOT NULL DEFAULT 0');
  addColumn('service_orders', 'product_label', 'TEXT');
  addColumn('clients', 'locality', 'TEXT');
  addColumn('service_orders', 'locality', 'TEXT');

  // Backfill free-text locality from the old localities table when present
  const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='localities'").get();
  if (tables) {
    if (columns('clients').includes('locality_id')) {
      db.exec(`
        UPDATE clients
        SET locality = (
          SELECT l.name FROM localities l WHERE l.id = clients.locality_id
        )
        WHERE (locality IS NULL OR locality = '') AND locality_id IS NOT NULL
      `);
    }
    if (columns('service_orders').includes('locality_id')) {
      db.exec(`
        UPDATE service_orders
        SET locality = (
          SELECT l.name FROM localities l WHERE l.id = service_orders.locality_id
        )
        WHERE (locality IS NULL OR locality = '') AND locality_id IS NOT NULL
      `);
    }
  }

  // Si ya había repuesto de catálogo, copiar el nombre al campo libre una sola vez
  db.exec(`
    UPDATE service_orders
    SET product_label = (
      SELECT p.name FROM products p WHERE p.id = service_orders.product_id
    )
    WHERE product_label IS NULL AND product_id IS NOT NULL
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS tariffs (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      scope TEXT NOT NULL CHECK (scope IN ('proveedor', 'tecnico', 'general')),
      provider_id INTEGER REFERENCES providers(id),
      technician_id INTEGER REFERENCES technicians(id),
      income_fixed REAL NOT NULL DEFAULT 0,
      income_per_hour REAL NOT NULL DEFAULT 0,
      income_per_km REAL NOT NULL DEFAULT 0,
      income_parts_pct REAL NOT NULL DEFAULT 0,
      cost_fixed REAL NOT NULL DEFAULT 0,
      cost_per_hour REAL NOT NULL DEFAULT 0,
      cost_per_km REAL NOT NULL DEFAULT 0,
      cost_parts_pct REAL NOT NULL DEFAULT 100,
      active INTEGER NOT NULL DEFAULT 1
    );
  `);

  const n = db.prepare('SELECT COUNT(*) AS n FROM tariffs').get().n;
  if (n === 0) {
    const providers = db.prepare('SELECT id, kind FROM providers ORDER BY id').all();
    const techs = db.prepare('SELECT id FROM technicians ORDER BY id').all();
    const ins = db.prepare(
      `INSERT INTO tariffs (
         name, scope, provider_id, technician_id,
         income_fixed, income_per_hour, income_per_km, income_parts_pct,
         cost_fixed, cost_per_hour, cost_per_km, cost_parts_pct, active
       ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`
    );
    ins.run('Tarifa general del taller', 'general', null, null, 10000, 0, 0, 0, 7000, 0, 0, 100);
    for (const p of providers) {
      if (p.kind === 'prestador') {
        ins.run('Contrato prestador (fijo + km)', 'proveedor', p.id, null, 18000, 0, 250, 0, 0, 0, 0, 0);
      } else {
        ins.run('Contrato proveedor (fijo + % repuestos)', 'proveedor', p.id, null, 12000, 0, 0, 20, 0, 0, 0, 0);
      }
    }
    if (techs[0]) ins.run('Costo técnico titular', 'tecnico', null, techs[0].id, 0, 0, 0, 0, 8000, 2500, 80, 100);
    if (techs[1]) ins.run('Costo técnico electrónica', 'tecnico', null, techs[1].id, 0, 0, 0, 0, 7500, 2400, 80, 100);
    if (techs[2]) ins.run('Costo técnico motos', 'tecnico', null, techs[2].id, 0, 0, 0, 0, 7000, 2200, 80, 100);
  }

  const billed = db.prepare('SELECT COUNT(*) AS n FROM service_orders WHERE hours > 0 OR km > 0').get().n;
  if (billed === 0) {
    db.exec(`
      UPDATE service_orders SET hours = 1.5, km = 18, parts_cost = 4200, parts_sale = 6800 WHERE id = 1;
      UPDATE service_orders SET hours = 2, km = 22, parts_cost = 0, parts_sale = 0 WHERE id = 2;
      UPDATE service_orders SET hours = 1, km = 35, parts_cost = 1500, parts_sale = 2500 WHERE id = 3;
      UPDATE service_orders SET hours = 1, km = 10, parts_cost = 8000, parts_sale = 11000 WHERE id = 4;
    `);
  }

  db.exec(`
    CREATE TABLE IF NOT EXISTS whatsapp_templates (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      body TEXT NOT NULL,
      status_id INTEGER REFERENCES service_statuses(id),
      auto_open INTEGER NOT NULL DEFAULT 0,
      active INTEGER NOT NULL DEFAULT 1
    );
  `);

  const wa = db.prepare('SELECT COUNT(*) AS n FROM whatsapp_templates').get().n;
  if (wa === 0) {
    const statuses = Object.fromEntries(
      db.prepare('SELECT id, name FROM service_statuses').all().map((s) => [s.name, s.id])
    );
    const ins = db.prepare(
      'INSERT INTO whatsapp_templates (name, body, status_id, auto_open, active) VALUES (?, ?, ?, ?, 1)'
    );
    ins.run(
      'Confirmación de visita',
      'Hola {cliente}, soy de Instal Service S.A. Te confirmamos la visita del {fecha} a las {hora} por “{trabajo}”. Técnico: {tecnico}. Ante dudas respondé este mensaje.',
      statuses['Ingresado'] || null,
      0
    );
    ins.run(
      'Esperando repuesto',
      'Hola {cliente}: tu orden #{orden} ({trabajo}) está en espera de repuesto. Te avisamos apenas avancemos.',
      statuses['Esperando repuesto'] || null,
      1
    );
    ins.run(
      'Listo para entregar',
      'Hola {cliente}: tu equipo ya está listo para entregar (orden #{orden}: {trabajo}). Coordinamos retiro o entrega?',
      statuses['Listo para entregar'] || null,
      1
    );
    ins.run(
      'Mensaje libre',
      'Hola {cliente}, te escribimos por tu servicio “{trabajo}” (orden #{orden}).',
      null,
      0
    );
  }
}
