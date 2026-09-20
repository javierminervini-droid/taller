import { mkdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { DatabaseSync } from 'node:sqlite';
import bcrypt from 'bcryptjs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const dataDir = join(__dirname, '..', 'data');
mkdirSync(dataDir, { recursive: true });

export const db = new DatabaseSync(join(dataDir, 'taller.db'));
db.exec('PRAGMA foreign_keys = ON');
db.exec(readFileSync(join(__dirname, 'schema.sql'), 'utf8'));

function count(table) {
  return db.prepare(`SELECT COUNT(*) AS n FROM ${table}`).get().n;
}

function insert(stmt, ...args) {
  const result = stmt.run(...args);
  return Number(result.lastInsertRowid);
}

export function seedIfEmpty() {
  if (count('users') > 0) return;

  const hash = bcrypt.hashSync('admin123', 10);
  const insertUser = db.prepare(
    'INSERT INTO users (username, password_hash, full_name, role) VALUES (?, ?, ?, ?)'
  );
  insert(insertUser, 'admin', hash, 'Administrador', 'admin');
  insert(insertUser, 'coord', bcrypt.hashSync('coord123', 10), 'Laura Coordinación', 'coordinador');
  const techUserId = insert(
    insertUser,
    'diego',
    bcrypt.hashSync('diego123', 10),
    'Diego Méndez',
    'tecnico'
  );
  const techUser2 = insert(
    insertUser,
    'sofia',
    bcrypt.hashSync('sofia123', 10),
    'Sofía Rivas',
    'tecnico'
  );

  const loc = db.prepare('INSERT INTO localities (name, province) VALUES (?, ?)');
  const locIds = {
    caba: insert(loc, 'CABA', 'Buenos Aires'),
    quilmes: insert(loc, 'Quilmes', 'Buenos Aires'),
    laPlata: insert(loc, 'La Plata', 'Buenos Aires'),
    moron: insert(loc, 'Morón', 'Buenos Aires'),
  };

  const prov = db.prepare(
    'INSERT INTO providers (name, kind, contact, notes) VALUES (?, ?, ?, ?)'
  );
  const proveedorId = insert(
    prov,
    'Repuestos del Sur',
    'proveedor',
    '11 4567-8900',
    'Base de clientes y stock de recambios'
  );
  const prestadorId = insert(
    prov,
    'Garantías Andinas',
    'prestador',
    '11 4789-1122',
    'Órdenes derivadas de garantía'
  );

  const tech = db.prepare(
    'INSERT INTO technicians (user_id, name, phone, specialty, active) VALUES (?, ?, ?, ?, 1)'
  );
  const t1 = insert(tech, techUserId, 'Diego Méndez', '11 5555-1001', 'Línea blanca');
  const t2 = insert(tech, techUser2, 'Sofía Rivas', '11 5555-1002', 'Electrónica');
  const t3 = insert(tech, null, 'Martín Acosta', '11 5555-1003', 'Motos y herramientas');

  const type = db.prepare('INSERT INTO product_types (name) VALUES (?)');
  const types = {
    lavarropas: insert(type, 'Lavarropas'),
    heladera: insert(type, 'Heladera'),
    aire: insert(type, 'Aire acondicionado'),
    moto: insert(type, 'Motocicleta'),
  };

  const prod = db.prepare(
    'INSERT INTO products (name, sku, product_type_id, provider_id, stock, price) VALUES (?, ?, ?, ?, ?, ?)'
  );
  const pCorrea = insert(prod, 'Correa lavarropas 1270 J5', 'COR-1270', types.lavarropas, proveedorId, 12, 18500);
  prod.run('Termostato heladera', 'TER-HL01', types.heladera, proveedorId, 8, 22100);
  prod.run('Capacitor 45uF', 'CAP-45', types.aire, proveedorId, 20, 9800);
  prod.run('Kit freno delantero', 'FRN-MOTO', types.moto, prestadorId, 4, 45200);

  const client = db.prepare(
    `INSERT INTO clients (provider_id, external_id, name, phone, email, locality_id, address)
     VALUES (?, ?, ?, ?, ?, ?, ?)`
  );
  const c1 = insert(client, proveedorId, 'P-1042', 'Ana López', '11 6001-2200', 'ana@correo.com', locIds.caba, 'Av. Rivadavia 2100');
  const c2 = insert(client, prestadorId, 'G-331', 'Carlos Pérez', '11 6002-1188', 'carlos@correo.com', locIds.quilmes, 'Calle Mitre 450');
  const c3 = insert(client, proveedorId, 'P-1188', 'María Suárez', '11 6003-4400', 'maria@correo.com', locIds.laPlata, 'Calle 12 n° 800');
  const c4 = insert(client, prestadorId, 'G-402', 'Jorge Díaz', '11 6004-9900', 'jorge@correo.com', locIds.moron, 'Belgrano 90');

  const st = db.prepare('INSERT INTO service_statuses (name, sort_order, color, is_closed) VALUES (?, ?, ?, ?)');
  const statuses = {
    ingresado: insert(st, 'Ingresado', 1, '#64748b', 0),
    diagnostico: insert(st, 'Diagnóstico', 2, '#2563eb', 0),
    espera: insert(st, 'Esperando repuesto', 3, '#d97706', 0),
    reparacion: insert(st, 'En reparación', 4, '#7c3aed', 0),
    listo: insert(st, 'Listo para entregar', 5, '#059669', 0),
    entregado: insert(st, 'Entregado', 6, '#334155', 1),
    cancelado: insert(st, 'Cancelado', 7, '#dc2626', 1),
  };

  const rule = db.prepare(
    `INSERT INTO followup_rules (name, trigger_type, status_id, hours, assign_role, active)
     VALUES (?, ?, ?, ?, ?, 1)`
  );
  rule.run('Avisar cliente: listo para entregar', 'status', statuses.listo, null, 'coordinador');
  rule.run('Gestionar compra de repuesto', 'status', statuses.espera, null, 'coordinador');
  rule.run('Diagnóstico demorado (48 h)', 'elapsed_hours', statuses.diagnostico, 48, 'coordinador');
  rule.run('Repuesto demorado (72 h)', 'elapsed_hours', statuses.espera, 72, 'coordinador');

  const iso = (offset) => {
    const d = new Date();
    d.setDate(d.getDate() + offset);
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  };

  const order = db.prepare(
    `INSERT INTO service_orders
      (client_id, technician_id, product_type_id, product_id, locality_id, provider_id,
       status_id, title, description, scheduled_date, scheduled_time, started_at, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`
  );

  const twoDaysAgo = new Date(Date.now() - 50 * 3600 * 1000).toISOString().replace('T', ' ').slice(0, 19);

  order.run(
    c1, t1, types.lavarropas, pCorrea, locIds.caba, proveedorId, statuses.espera,
    'Cambio de correa', 'No centrifuga. Cliente de Repuestos del Sur.',
    iso(0), '09:00', twoDaysAgo, twoDaysAgo
  );
  order.run(
    c2, t2, types.heladera, null, locIds.quilmes, prestadorId, statuses.diagnostico,
    'Heladera no enfría', 'Orden de garantía G-331.',
    iso(0), '11:30', twoDaysAgo, twoDaysAgo
  );
  order.run(
    c3, t1, types.aire, null, locIds.laPlata, proveedorId, statuses.listo,
    'Carga de gas y limpieza', 'Unidad lista. Coordinar entrega.',
    iso(0), '15:00', iso(-1) + ' 10:00:00', iso(-2) + ' 09:00:00'
  );
  order.run(
    c4, t3, types.moto, null, locIds.moron, prestadorId, statuses.ingresado,
    'Frenos delanteros', 'Prestador derivó inspección.',
    iso(1), '10:00', null, iso(0) + ' 08:00:00'
  );
}
