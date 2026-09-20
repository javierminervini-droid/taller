/**
 * QA smoke tests for Taller Gestión Go API (phase 1)
 * Usage: node scripts/qa.mjs
 * Optional: QA_BASE=http://127.0.0.1:3847
 */
const BASE = process.env.QA_BASE || 'http://127.0.0.1:3847';

let passed = 0;
let failed = 0;

function ok(name, detail = '') {
  passed += 1;
  console.log(`  ✓ ${name}${detail ? ` — ${detail}` : ''}`);
}
function fail(name, detail = '') {
  failed += 1;
  console.log(`  ✗ ${name}${detail ? ` — ${detail}` : ''}`);
}
function skip(name, detail = '') {
  console.log(`  ○ ${name}${detail ? ` — ${detail}` : ''} (fase 2)`);
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
  console.log(`\nQA Instal Service S.A. (Go) → ${BASE}\n`);

  {
    const { res, data } = await req('/health');
    if (res.ok && data.ok) ok('Health');
    else fail('Health', `status ${res.status}`);
  }

  {
    const { res } = await req('/api/auth/login', { method: 'POST', body: { username: 'admin', password: 'wrong' } });
    if (res.status === 401) ok('Login inválido rechazado');
    else fail('Login inválido', `status ${res.status}`);
  }

  let adminToken;
  {
    const { res, data } = await req('/api/auth/login', { method: 'POST', body: { username: 'admin', password: 'admin123' } });
    if (res.ok && data.token && data.user?.role === 'admin') {
      adminToken = data.token;
      ok('Login admin', data.user.full_name);
    } else fail('Login admin', JSON.stringify(data));
  }

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

  {
    const { res, data } = await req('/api/lookups', { token: adminToken });
    if (res.ok && data.statuses?.length && (data.unitTypes?.length || data.productTypes?.length)) {
      ok('Lookups admin', `${data.statuses.length} estados`);
    } else fail('Lookups admin');
  }

  let orderId;
  {
    const { res, data } = await req('/api/agenda', { token: adminToken });
    const rows = data.requests || data.orders || [];
    if (res.ok && Array.isArray(rows) && rows.length > 0) {
      orderId = rows[0].id;
      const cols = rows[0];
      const hasLoad = cols.load_date || cols.received_at || cols.client_created_at;
      if (hasLoad) ok('Agenda admin con solicitudes', `${rows.length} filas, date=${data.date}`);
      else fail('Agenda falta load_date/received_at');
    } else fail('Agenda admin', `requests=${rows.length}`);
  }

  {
    const { res, data } = await req('/api/agenda', { token: techToken });
    const rows = data.requests || data.orders || [];
    if (res.ok && data.technicians?.length === 1) {
      const foreign = rows.some((o) => o.technician_id !== data.technicians[0].id);
      if (!foreign) ok('Agenda técnico solo sus solicitudes', `${rows.length} filas`);
      else fail('Agenda técnico ve solicitudes ajenas');
    } else fail('Agenda técnico');
  }

  {
    const { res } = await req('/api/users', { token: techToken });
    if (res.status === 403) ok('Técnico sin acceso a usuarios');
    else fail('Técnico usuarios', `status ${res.status}`);
  }

  {
    const { res, data } = await req('/api/clients', { token: adminToken });
    if (res.ok && data.length >= 1) ok('Listado clientes', `${data.length}`);
    else fail('Clientes');
  }

  {
    const { res, data } = await req('/api/providers', { token: adminToken });
    if (res.ok && Array.isArray(data) && data.length >= 1) ok('Listado proveedores', `${data.length}`);
    else fail('Proveedores', JSON.stringify(data).slice(0, 120));
  }

  {
    const y = new Date().getFullYear();
    const { res, data } = await req(`/api/service-requests?year=${y}`, { token: adminToken });
    if (res.ok && Array.isArray(data.requests || data.orders) && Array.isArray(data.years)) {
      ok('Histórico solicitudes por año', `${data.total} en ${y}`);
    } else fail('Histórico solicitudes', JSON.stringify(data).slice(0, 120));
  }

  if (orderId) {
    const { res, data } = await req(`/api/service-requests/${orderId}`, {
      method: 'PATCH',
      token: adminToken,
      body: { appliance_model: 'Lavarropas QA Demo' },
    });
    if (res.ok && data.ok) ok('PATCH modelo libre');
    else fail('PATCH modelo', JSON.stringify(data));
  }

  {
    const { res } = await req('/api/followups', { token: techToken });
    if (res.status === 403) ok('Seguimientos bloqueados a técnico');
    else fail('Seguimientos técnico', `status ${res.status}`);
  }

  {
    const { res, data } = await req('/api/followups', { token: adminToken });
    if (res.ok && Array.isArray(data)) ok('Seguimientos admin', `${data.length}`);
    else fail('Seguimientos admin', JSON.stringify(data).slice(0, 120));
  }

  // Phase 2 — live on Go API
  {
    const { res, data } = await req('/api/whatsapp/templates', { token: adminToken });
    if (res.ok && Array.isArray(data)) ok('WhatsApp templates', `${data.length}`);
    else fail('WhatsApp templates', JSON.stringify(data).slice(0, 120));
  }
  {
    const { res, data } = await req('/api/tariffs', { token: adminToken });
    if (res.ok && Array.isArray(data)) ok('Tarifario', `${data.length}`);
    else fail('Tarifario', JSON.stringify(data).slice(0, 120));
  }
  {
    const res = await req('/api/clients/template.xlsx', { token: adminToken, raw: true });
    if (res.ok) ok('Plantilla Excel', res.headers.get('content-type') || 'ok');
    else fail('Plantilla Excel', `status ${res.status}`);
  }

  console.log(`\nResultado: ${passed} ok, ${failed} fail\n`);
  process.exit(failed ? 1 : 0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
