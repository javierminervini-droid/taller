/**
 * QA smoke tests for Taller Gestión API
 * Usage: node scripts/qa.mjs
 */
const BASE = process.env.QA_BASE || 'http://127.0.0.1:3847';

let passed = 0;
let failed = 0;
const results = [];

function ok(name, detail = '') {
  passed += 1;
  results.push({ ok: true, name, detail });
  console.log(`  ✓ ${name}${detail ? ` — ${detail}` : ''}`);
}
function fail(name, detail = '') {
  failed += 1;
  results.push({ ok: false, name, detail });
  console.log(`  ✗ ${name}${detail ? ` — ${detail}` : ''}`);
}

async function req(path, { method = 'GET', token, body, raw } = {}) {
  const headers = {};
  if (token) headers.Authorization = `Bearer ${token}`;
  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
    body = JSON.stringify(body);
  }
  const res = await fetch(`${BASE}${path}`, { method, headers, body });
  if (raw) return res;
  const data = await res.json().catch(() => ({}));
  return { res, data };
}

async function main() {
  console.log(`\nQA Instal Service S.A. → ${BASE}\n`);

  // Health / UI
  {
    const res = await fetch(`${BASE}/`);
    if (res.ok && (await res.text()).includes('Instal Service')) ok('UI index.html sirve login');
    else fail('UI index.html', `status ${res.status}`);
  }

  // Bad login
  {
    const { res } = await req('/api/auth/login', { method: 'POST', body: { username: 'admin', password: 'wrong' } });
    if (res.status === 401) ok('Login inválido rechazado');
    else fail('Login inválido', `status ${res.status}`);
  }

  // Admin login
  let adminToken;
  {
    const { res, data } = await req('/api/auth/login', { method: 'POST', body: { username: 'admin', password: 'admin123' } });
    if (res.ok && data.token && data.user?.role === 'admin') {
      adminToken = data.token;
      ok('Login admin', data.user.full_name);
    } else fail('Login admin', JSON.stringify(data));
  }

  // Tech login
  let techToken;
  {
    const { res, data } = await req('/api/auth/login', { method: 'POST', body: { username: 'diego', password: 'diego123' } });
    if (res.ok && data.token && data.user?.role === 'tecnico' && data.user.technician_id) {
      techToken = data.token;
      ok('Login técnico', `tech_id=${data.user.technician_id}`);
    } else fail('Login técnico', JSON.stringify(data));
  }

  if (!adminToken || !techToken) {
    console.log(`\nAbortado: ${passed} ok, ${failed} fail\n`);
    process.exit(1);
  }

  // Lookups
  {
    const { res, data } = await req('/api/lookups', { token: adminToken });
    if (res.ok && data.statuses?.length && data.productTypes?.length) ok('Lookups admin', `${data.statuses.length} estados`);
    else fail('Lookups admin');
  }

  // Agenda admin
  let orderId;
  {
    const { res, data } = await req('/api/agenda', { token: adminToken });
    if (res.ok && Array.isArray(data.orders) && data.orders.length > 0) {
      orderId = data.orders[0].id;
      const cols = data.orders[0];
      const hasLoad = cols.load_date || cols.client_created_at;
      if (hasLoad) ok('Agenda admin con órdenes y fecha carga', `${data.orders.length} filas, date=${data.date}`);
      else fail('Agenda falta load_date');
    } else fail('Agenda admin', `orders=${data.orders?.length}`);
  }

  // Agenda tech scoped
  {
    const { res, data } = await req('/api/agenda', { token: techToken });
    if (res.ok && data.technicians?.length === 1) {
      const foreign = data.orders.some((o) => o.technician_id !== data.technicians[0].id);
      if (!foreign) ok('Agenda técnico solo sus órdenes', `${data.orders.length} filas`);
      else fail('Agenda técnico ve órdenes ajenas');
    } else fail('Agenda técnico');
  }

  // Tech cannot list users
  {
    const { res } = await req('/api/users', { token: techToken });
    if (res.status === 403) ok('Técnico sin acceso a usuarios');
    else fail('Técnico usuarios', `status ${res.status}`);
  }

  // Clients
  {
    const { res, data } = await req('/api/clients', { token: adminToken });
    if (res.ok && data.length >= 1) ok('Listado clientes', `${data.length}`);
    else fail('Clientes');
  }

  // WhatsApp templates + message URL
  {
    const { res, data } = await req('/api/whatsapp/templates', { token: adminToken });
    if (res.ok && data.length >= 1) ok('Plantillas WhatsApp', `${data.length}`);
    else fail('Plantillas WhatsApp');
  }
  if (orderId) {
    const { res, data } = await req(`/api/orders/${orderId}/whatsapp`, { token: adminToken });
    if (res.ok && data.url?.startsWith('https://wa.me/')) ok('WhatsApp URL de orden', data.template);
    else fail('WhatsApp URL', JSON.stringify(data));
  }

  // Tariffs + results
  {
    const { res, data } = await req('/api/tariffs', { token: adminToken });
    if (res.ok && data.length >= 1) ok('Tarifario', `${data.length} tarifas`);
    else fail('Tarifario');
  }
  {
    const { res, data } = await req('/api/results?group=proveedor', { token: adminToken });
    if (res.ok && data.totals && Array.isArray(data.rows)) ok('Ganancias por prestador', `profit=${data.totals.profit}`);
    else fail('Ganancias');
  }

  // Excel template
  {
    const res = await req('/api/clients/template.xlsx', { token: adminToken, raw: true });
    const ct = res.headers.get('content-type') || '';
    if (res.ok && ct.includes('spreadsheet')) ok('Plantilla Excel Salesforce');
    else fail('Plantilla Excel', `status ${res.status} ct=${ct}`);
  }

  // Historical orders
  {
    const y = new Date().getFullYear();
    const { res, data } = await req(`/api/orders?year=${y}`, { token: adminToken });
    if (res.ok && Array.isArray(data.orders) && Array.isArray(data.years)) {
      ok('Histórico órdenes por año', `${data.total} en ${y}`);
    } else fail('Histórico órdenes', JSON.stringify(data).slice(0, 120));
  }

  // Patch product_label
  if (orderId) {
    const { res, data } = await req(`/api/orders/${orderId}`, {
      method: 'PATCH',
      token: adminToken,
      body: { product_label: 'Lavarropas QA Demo' },
    });
    if (res.ok && data.ok) ok('PATCH producto libre');
    else fail('PATCH producto', JSON.stringify(data));
  }

  // Followups blocked for tech
  {
    const { res } = await req('/api/followups', { token: techToken });
    if (res.status === 403) ok('Seguimientos bloqueados a técnico');
    else fail('Seguimientos técnico', `status ${res.status}`);
  }

  console.log(`\nResultado: ${passed} ok, ${failed} fail\n`);
  process.exit(failed ? 1 : 0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
