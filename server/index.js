import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import Fastify from 'fastify';
import cors from '@fastify/cors';
import fastifyStatic from '@fastify/static';
import bcrypt from 'bcryptjs';
import { db, seedIfEmpty } from './db.js';
import { canManage, login, publicUser, requireAuth, requireRole, signUser, technicianOf } from './auth.js';
import { generateFollowups, listFollowups } from './followups.js';
import { migrate } from './migrate.js';
import { buildTemplateBuffer, importClientsFromBuffer } from './import-clients.js';
import { listTariffs, resultsReport } from './finance.js';
import { listWhatsAppTemplates, messageForOrder, telHref } from './whatsapp.js';
import multipart from '@fastify/multipart';

seedIfEmpty();
migrate();

const app = Fastify({ logger: false });
const PORT = Number(process.env.PORT || 3847);
const __dirname = dirname(fileURLToPath(import.meta.url));

await app.register(cors, { origin: true });
await app.register(multipart, { limits: { fileSize: 8 * 1024 * 1024 } });

app.addHook('preHandler', async (request, reply) => {
  if (request.url.startsWith('/api/auth/login')) return;
  if (!request.url.startsWith('/api/')) return;
  const user = requireAuth(request, reply);
  if (!user) return reply;
});

app.post('/api/auth/login', async (request, reply) => {
  const { username, password } = request.body || {};
  const user = login(username, password);
  if (!user) return reply.code(401).send({ error: 'Usuario o clave incorrectos' });
  return {
    token: signUser(user),
    user: publicUser(user),
  };
});

app.get('/api/auth/me', async (request) => {
  const row = db.prepare('SELECT * FROM users WHERE id = ?').get(request.user.id);
  if (!row) return { user: request.user };
  return { user: publicUser(row) };
});

app.get('/api/lookups', async (request) => {
  const statuses = db.prepare('SELECT * FROM service_statuses ORDER BY sort_order').all();
  const productTypes = db.prepare('SELECT * FROM product_types ORDER BY name').all();
  if (request.user.role === 'tecnico') {
    const tech = technicianOf(request.user);
    return {
      technicians: tech ? [tech] : [],
      providers: [],
      productTypes,
      statuses,
      products: db.prepare('SELECT * FROM products ORDER BY name').all(),
      users: [],
    };
  }
  return {
    technicians: db.prepare('SELECT * FROM technicians WHERE active = 1 ORDER BY name').all(),
    providers: db.prepare('SELECT * FROM providers ORDER BY name').all(),
    productTypes,
    statuses,
    products: db.prepare('SELECT * FROM products ORDER BY name').all(),
    users: db.prepare('SELECT id, username, full_name, role, active FROM users ORDER BY full_name').all(),
  };
});

function localDate(value) {
  const d = value ? new Date(value) : new Date();
  if (Number.isNaN(d.getTime())) {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
  }
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function orderFilters(query) {
  const clauses = [];
  const params = [];
  if (query.date) {
    clauses.push('o.scheduled_date = ?');
    params.push(query.date);
  }
  if (query.year) {
    clauses.push(`strftime('%Y', o.created_at) = ?`);
    params.push(String(query.year));
  }
  if (query.month) {
    clauses.push(`strftime('%m', o.created_at) = ?`);
    params.push(String(query.month).padStart(2, '0'));
  }
  if (query.technicianId) {
    clauses.push('o.technician_id = ?');
    params.push(Number(query.technicianId));
  }
  if (query.productTypeId) {
    clauses.push('o.product_type_id = ?');
    params.push(Number(query.productTypeId));
  }
  if (query.providerId) {
    clauses.push('o.provider_id = ?');
    params.push(Number(query.providerId));
  }
  if (query.locality) {
    clauses.push('o.locality LIKE ?');
    params.push(`%${query.locality}%`);
  }
  if (query.statusId) {
    clauses.push('o.status_id = ?');
    params.push(Number(query.statusId));
  }
  if (query.clientId) {
    clauses.push('o.client_id = ?');
    params.push(Number(query.clientId));
  }
  return { where: clauses.length ? `WHERE ${clauses.join(' AND ')}` : '', params };
}

const ORDER_SELECT = `
  SELECT o.*,
         c.name AS client_name, c.phone AS client_phone, c.email AS client_email,
         c.address AS client_address, c.created_at AS client_created_at, c.source AS client_source,
         t.name AS technician_name,
         pt.name AS product_type_name,
         p.name AS product_name,
         pr.name AS provider_name, pr.kind AS provider_kind,
         s.name AS status_name, s.color AS status_color,
         date(COALESCE(c.created_at, o.created_at)) AS load_date,
         date(o.created_at) AS entry_date,
         strftime('%Y', o.created_at) AS entry_year,
         strftime('%m', o.created_at) AS entry_month
  FROM service_orders o
  JOIN clients c ON c.id = o.client_id
  JOIN service_statuses s ON s.id = o.status_id
  LEFT JOIN technicians t ON t.id = o.technician_id
  LEFT JOIN product_types pt ON pt.id = o.product_type_id
  LEFT JOIN products p ON p.id = o.product_id
  LEFT JOIN providers pr ON pr.id = o.provider_id
`;

function scopedQuery(user, query) {
  const next = { ...query };
  if (user.role === 'tecnico') {
    const tech = technicianOf(user);
    next.technicianId = tech ? String(tech.id) : '-1';
  }
  return next;
}

app.get('/api/agenda', async (request, reply) => {
  if (request.user.role === 'tecnico' && !technicianOf(request.user)) {
    return reply.code(403).send({ error: 'Tu usuario no está vinculado a un perfil de técnico' });
  }
  const q = scopedQuery(request.user, request.query);
  const date = q.date || localDate();
  const { where, params } = orderFilters({ ...q, date });
  generateFollowups();
  const orders = db.prepare(`${ORDER_SELECT} ${where} ORDER BY o.scheduled_time, o.id`).all(...params);
  const technicians = request.user.role === 'tecnico'
    ? [technicianOf(request.user)].filter(Boolean)
    : db.prepare('SELECT * FROM technicians WHERE active = 1 ORDER BY name').all();
  return { date, technicians, orders };
});

app.get('/api/orders', async (request) => {
  const q = scopedQuery(request.user, request.query);
  const { where, params } = orderFilters(q);
  const orders = db.prepare(
    `${ORDER_SELECT} ${where} ORDER BY o.created_at DESC, o.id DESC`
  ).all(...params);
  const years = db.prepare(
    `SELECT DISTINCT strftime('%Y', created_at) AS year
     FROM service_orders
     WHERE created_at IS NOT NULL
     ORDER BY year DESC`
  ).all().map((r) => r.year).filter(Boolean);
  return {
    orders,
    years,
    year: q.year || null,
    month: q.month || null,
    total: orders.length,
  };
});

app.get('/api/orders/:id', async (request, reply) => {
  const order = db.prepare(`${ORDER_SELECT} WHERE o.id = ?`).get(Number(request.params.id));
  if (!order) return reply.code(404).send({ error: 'Orden no encontrada' });
  if (request.user.role === 'tecnico') {
    const tech = technicianOf(request.user);
    if (!tech || order.technician_id !== tech.id) {
      return reply.code(403).send({ error: 'Solo podés ver tus órdenes' });
    }
  }
  return order;
});

app.post('/api/orders', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    `INSERT INTO service_orders
      (client_id, technician_id, product_type_id, product_id, product_label, locality, provider_id,
       status_id, title, description, scheduled_date, scheduled_time,
       hours, km, parts_cost, parts_sale, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`
  ).run(
    b.client_id, b.technician_id || null, b.product_type_id || null, b.product_id || null,
    b.product_label || null,
    b.locality || null, b.provider_id || null, b.status_id,
    b.title || b.product_label || 'Servicio', b.description || null,
    b.scheduled_date || null, b.scheduled_time || null,
    Number(b.hours || 0), Number(b.km || 0), Number(b.parts_cost || 0), Number(b.parts_sale || 0)
  );
  generateFollowups();
  return { id: Number(result.lastInsertRowid) };
});

app.patch('/api/orders/:id', async (request, reply) => {
  const id = Number(request.params.id);
  const current = db.prepare('SELECT * FROM service_orders WHERE id = ?').get(id);
  if (!current) return reply.code(404).send({ error: 'Orden no encontrada' });
  if (request.user.role === 'tecnico') {
    const tech = technicianOf(request.user);
    if (!tech || tech.id !== current.technician_id) {
      return reply.code(403).send({ error: 'Solo podés actualizar tus órdenes' });
    }
  }
  const b = request.body || {};
  const allowed = request.user.role === 'tecnico'
    ? {
        title: b.title ?? current.title,
        description: b.description ?? current.description,
        status_id: b.status_id ?? current.status_id,
        product_id: current.product_id,
        product_label: b.product_label ?? current.product_label,
        product_type_id: b.product_type_id ?? current.product_type_id,
        scheduled_time: b.scheduled_time ?? current.scheduled_time,
        scheduled_date: current.scheduled_date,
        client_id: current.client_id,
        technician_id: current.technician_id,
        locality: b.locality ?? current.locality,
        provider_id: current.provider_id,
        hours: current.hours,
        km: current.km,
        parts_cost: current.parts_cost,
        parts_sale: current.parts_sale,
      }
    : { ...current, ...b };
  let startedAt = current.started_at;
  let completedAt = current.completed_at;
  if (allowed.status_id && Number(allowed.status_id) !== current.status_id) {
    const st = db.prepare('SELECT * FROM service_statuses WHERE id = ?').get(Number(allowed.status_id));
    if (!startedAt) startedAt = new Date().toISOString().replace('T', ' ').slice(0, 19);
    if (st?.is_closed) completedAt = new Date().toISOString().replace('T', ' ').slice(0, 19);
  }
  db.prepare(
    `UPDATE service_orders SET
      client_id=?, technician_id=?, product_type_id=?, product_id=?, product_label=?, locality=?, provider_id=?,
      status_id=?, title=?, description=?, scheduled_date=?, scheduled_time=?,
      hours=?, km=?, parts_cost=?, parts_sale=?,
      started_at=?, completed_at=?, updated_at=datetime('now')
     WHERE id=?`
  ).run(
    allowed.client_id, allowed.technician_id || null, allowed.product_type_id || null, allowed.product_id || null,
    allowed.product_label || null,
    allowed.locality || null, allowed.provider_id || null, allowed.status_id, allowed.title, allowed.description || null,
    allowed.scheduled_date || null, allowed.scheduled_time || null,
    Number(allowed.hours || 0), Number(allowed.km || 0), Number(allowed.parts_cost || 0), Number(allowed.parts_sale || 0),
    startedAt, completedAt, id
  );
  generateFollowups();
  const statusChanged = allowed.status_id && Number(allowed.status_id) !== current.status_id;
  let whatsapp = null;
  if (statusChanged) {
    const order = db.prepare(`${ORDER_SELECT} WHERE o.id = ?`).get(id);
    if (order) whatsapp = messageForOrder(order);
  }
  return { ok: true, whatsapp: whatsapp?.url ? whatsapp : null };
});

app.get('/api/clients', async (request) => {
  const providerId = request.query.providerId;
  const source = request.query.source;
  if (request.user.role === 'tecnico') {
    const tech = technicianOf(request.user);
    return db.prepare(
      `SELECT DISTINCT c.*, p.name AS provider_name, p.kind AS provider_kind
       FROM clients c
       JOIN service_orders o ON o.client_id = c.id
       LEFT JOIN providers p ON p.id = c.provider_id
       WHERE o.technician_id = ?
       ORDER BY c.name`
    ).all(tech?.id || -1);
  }
  const clauses = [];
  const params = [];
  if (providerId === 'own') clauses.push('c.provider_id IS NULL');
  else if (providerId) {
    clauses.push('c.provider_id = ?');
    params.push(Number(providerId));
  }
  if (source) {
    clauses.push('c.source = ?');
    params.push(source);
  }
  const where = clauses.length ? `WHERE ${clauses.join(' AND ')}` : '';
  return db.prepare(
    `SELECT c.*, p.name AS provider_name, p.kind AS provider_kind
     FROM clients c
     LEFT JOIN providers p ON p.id = c.provider_id
     ${where}
     ORDER BY c.name`
  ).all(...params);
});

app.get('/api/clients/template.xlsx', async (request, reply) => {
  if (!canManage(request.user)) return reply.code(403).send({ error: 'Sin permiso para esta acción' });
  const buf = buildTemplateBuffer();
  return reply
    .header('Content-Type', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
    .header('Content-Disposition', 'attachment; filename="plantilla-clientes-salesforce.xlsx"')
    .send(buf);
});

app.post('/api/clients/import', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const file = await request.file();
  if (!file) return reply.code(400).send({ error: 'Adjuntá el Excel de Salesforce' });
  const providerId = Number(file.fields?.provider_id?.value || request.query.providerId);
  if (!providerId) return reply.code(400).send({ error: 'Elegí el prestador/proveedor de origen' });
  const buffer = await file.toBuffer();
  try {
    return importClientsFromBuffer(buffer, providerId);
  } catch {
    return reply.code(400).send({ error: 'No se pudo leer el Excel. Usá la plantilla del sistema.' });
  }
});

app.post('/api/clients', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    `INSERT INTO clients (provider_id, external_id, name, phone, email, locality, address, notes, source)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'manual')`
  ).run(b.provider_id || null, b.external_id || null, b.name, b.phone || null, b.email || null,
    b.locality || null, b.address || null, b.notes || null);
  return { id: Number(result.lastInsertRowid) };
});

app.patch('/api/clients/:id', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const id = Number(request.params.id);
  const current = db.prepare('SELECT * FROM clients WHERE id = ?').get(id);
  if (!current) return reply.code(404).send({ error: 'Cliente no encontrado' });
  const b = { ...current, ...request.body };
  db.prepare(
    `UPDATE clients SET provider_id=?, external_id=?, name=?, phone=?, email=?, locality=?, address=?, notes=?
     WHERE id=?`
  ).run(
    b.provider_id || null, b.external_id || null, b.name, b.phone || null, b.email || null,
    b.locality || null, b.address || null, b.notes || null, id
  );
  return { ok: true };
});
app.post('/api/providers', async (request, reply) => {
  if (!requireRole(request.user, ['admin'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    'INSERT INTO providers (name, kind, contact, notes) VALUES (?, ?, ?, ?)'
  ).run(b.name, b.kind, b.contact || null, b.notes || null);
  return { id: Number(result.lastInsertRowid) };
});

app.get('/api/technicians', async () => db.prepare('SELECT * FROM technicians ORDER BY name').all());
app.post('/api/technicians', async (request, reply) => {
  if (!requireRole(request.user, ['admin'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    'INSERT INTO technicians (user_id, name, phone, specialty, active) VALUES (?, ?, ?, ?, 1)'
  ).run(b.user_id || null, b.name, b.phone || null, b.specialty || null);
  return { id: Number(result.lastInsertRowid) };
});

app.get('/api/products', async () =>
  db.prepare(
    `SELECT p.*, pt.name AS type_name, pr.name AS provider_name
     FROM products p
     LEFT JOIN product_types pt ON pt.id = p.product_type_id
     LEFT JOIN providers pr ON pr.id = p.provider_id
     ORDER BY p.name`
  ).all()
);
app.post('/api/products', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    `INSERT INTO products (name, sku, product_type_id, provider_id, stock, price)
     VALUES (?, ?, ?, ?, ?, ?)`
  ).run(b.name, b.sku || null, b.product_type_id || null, b.provider_id || null, b.stock || 0, b.price || 0);
  return { id: Number(result.lastInsertRowid) };
});

app.post('/api/product-types', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const result = db.prepare('INSERT INTO product_types (name) VALUES (?)').run(request.body.name);
  return { id: Number(result.lastInsertRowid) };
});
app.patch('/api/product-types/:id', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  db.prepare('UPDATE product_types SET name = ? WHERE id = ?').run(request.body.name, Number(request.params.id));
  return { ok: true };
});

app.get('/api/followups', async (request, reply) => {
  if (request.user.role === 'tecnico') {
    return reply.code(403).send({ error: 'Los seguimientos los gestiona administración' });
  }
  return listFollowups(request.query.status || 'pendiente');
});
app.patch('/api/followups/:id', async (request) => {
  const id = Number(request.params.id);
  const status = request.body.status === 'hecho' ? 'hecho' : 'pendiente';
  db.prepare(
    `UPDATE followups SET status = ?, completed_at = CASE WHEN ? = 'hecho' THEN datetime('now') ELSE NULL END
     WHERE id = ?`
  ).run(status, status, id);
  return { ok: true };
});

app.get('/api/followup-rules', async (request, reply) => {
  if (!canManage(request.user)) return reply.code(403).send({ error: 'Sin permiso para esta acción' });
  return db.prepare(
    `SELECT r.*, s.name AS status_name FROM followup_rules r
     LEFT JOIN service_statuses s ON s.id = r.status_id
     ORDER BY r.id`
  ).all();
});
app.post('/api/followup-rules', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    `INSERT INTO followup_rules (name, trigger_type, status_id, hours, assign_role, active)
     VALUES (?, ?, ?, ?, ?, 1)`
  ).run(b.name, b.trigger_type, b.status_id || null, b.hours || null, b.assign_role || 'coordinador');
  return { id: Number(result.lastInsertRowid) };
});

app.get('/api/users', async (request, reply) => {
  if (!requireRole(request.user, ['admin'], reply)) return reply;
  return db.prepare('SELECT id, username, full_name, role, active FROM users ORDER BY id').all();
});
app.post('/api/users', async (request, reply) => {
  if (!requireRole(request.user, ['admin'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    'INSERT INTO users (username, password_hash, full_name, role, active) VALUES (?, ?, ?, ?, 1)'
  ).run(b.username, bcrypt.hashSync(b.password, 10), b.full_name, b.role);
  const userId = Number(result.lastInsertRowid);
  if (b.role === 'tecnico') {
    db.prepare(
      'INSERT INTO technicians (user_id, name, phone, specialty, active) VALUES (?, ?, ?, ?, 1)'
    ).run(userId, b.full_name, b.phone || null, b.specialty || null);
  }
  return { id: userId };
});

app.get('/api/whatsapp/templates', async (request, reply) => {
  if (!canManage(request.user)) return reply.code(403).send({ error: 'Sin permiso para esta acción' });
  return listWhatsAppTemplates();
});
app.post('/api/whatsapp/templates', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    `INSERT INTO whatsapp_templates (name, body, status_id, auto_open, active)
     VALUES (?, ?, ?, ?, 1)`
  ).run(b.name, b.body, b.status_id || null, b.auto_open ? 1 : 0);
  return { id: Number(result.lastInsertRowid) };
});
app.patch('/api/whatsapp/templates/:id', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const id = Number(request.params.id);
  const current = db.prepare('SELECT * FROM whatsapp_templates WHERE id = ?').get(id);
  if (!current) return reply.code(404).send({ error: 'Plantilla no encontrada' });
  const b = { ...current, ...request.body };
  db.prepare(
    `UPDATE whatsapp_templates SET name=?, body=?, status_id=?, auto_open=?, active=? WHERE id=?`
  ).run(b.name, b.body, b.status_id || null, b.auto_open ? 1 : 0, b.active === 0 ? 0 : 1, id);
  return { ok: true };
});

app.get('/api/orders/:id/whatsapp', async (request, reply) => {
  const order = db.prepare(`${ORDER_SELECT} WHERE o.id = ?`).get(Number(request.params.id));
  if (!order) return reply.code(404).send({ error: 'Orden no encontrada' });
  if (request.user.role === 'tecnico') {
    const tech = technicianOf(request.user);
    if (!tech || order.technician_id !== tech.id) {
      return reply.code(403).send({ error: 'Solo podés contactar clientes de tus órdenes' });
    }
  }
  const templateId = request.query.templateId ? Number(request.query.templateId) : null;
  const msg = messageForOrder(order, templateId);
  if (!msg || msg.error) {
    return reply.code(400).send({ error: msg?.error || 'No hay plantilla o teléfono' });
  }
  return { ...msg, tel: telHref(order.client_phone) };
});

app.get('/api/tariffs', async (request, reply) => {
  if (!canManage(request.user)) return reply.code(403).send({ error: 'Sin permiso para esta acción' });
  return listTariffs();
});
app.post('/api/tariffs', async (request, reply) => {
  if (!requireRole(request.user, ['admin', 'coordinador'], reply)) return reply;
  const b = request.body;
  const result = db.prepare(
    `INSERT INTO tariffs (
       name, scope, provider_id, technician_id,
       income_fixed, income_per_hour, income_per_km, income_parts_pct,
       cost_fixed, cost_per_hour, cost_per_km, cost_parts_pct, active
     ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`
  ).run(
    b.name, b.scope, b.provider_id || null, b.technician_id || null,
    Number(b.income_fixed || 0), Number(b.income_per_hour || 0), Number(b.income_per_km || 0), Number(b.income_parts_pct || 0),
    Number(b.cost_fixed || 0), Number(b.cost_per_hour || 0), Number(b.cost_per_km || 0), Number(b.cost_parts_pct ?? 100)
  );
  return { id: Number(result.lastInsertRowid) };
});

app.get('/api/results', async (request, reply) => {
  if (!canManage(request.user)) return reply.code(403).send({ error: 'Sin permiso para esta acción' });
  const { from, to, group } = request.query;
  return resultsReport({ from, to, group: group || 'proveedor' });
});

const dist = join(__dirname, '..', 'client', 'dist');
if (existsSync(dist)) {
  await app.register(fastifyStatic, { root: dist });
  app.setNotFoundHandler((request, reply) => {
    if (request.url.startsWith('/api/')) return reply.code(404).send({ error: 'No encontrado' });
    return reply.sendFile('index.html');
  });
}

try {
  await app.listen({ port: PORT, host: '0.0.0.0' });
  console.log(`Instal Service S.A. — gestión de taller en http://localhost:${PORT}`);
  console.log('Usuarios demo: admin/admin123  coord/coord123  diego/diego123');
} catch (err) {
  console.error(err);
  process.exit(1);
}
