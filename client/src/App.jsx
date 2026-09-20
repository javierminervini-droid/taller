import { useEffect, useState } from 'react';
import { api, downloadFile, getToken, setToken } from './api.js';

const today = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
};

const ROLE_LABEL = {
  admin: 'Administrador',
  coordinador: 'Coordinación',
  tecnico: 'Técnico',
};

const PAGES = [
  ['agenda', 'Agenda', ['admin', 'coordinador', 'tecnico']],
  ['seguimientos', 'Seguimientos', ['admin', 'coordinador']],
  ['ordenes', 'Histórico', ['admin', 'coordinador']],
  ['clientes', 'Clientes', ['admin', 'coordinador']],
  ['tarifas', 'Tarifario', ['admin', 'coordinador']],
  ['resultados', 'Ganancias', ['admin', 'coordinador']],
  ['catalogo', 'Repuestos', ['admin', 'coordinador']],
  ['tecnicos', 'Técnicos', ['admin']],
  ['proveedores', 'Proveedores', ['admin']],
  ['reglas', 'Reglas', ['admin', 'coordinador']],
  ['whatsapp', 'WhatsApp', ['admin', 'coordinador']],
  ['usuarios', 'Usuarios', ['admin']],
];

function openWhatsApp(url) {
  if (!url) return;
  window.open(url, '_blank', 'noopener,noreferrer');
}

function callClient(phone) {
  const digits = String(phone || '').replace(/\D/g, '');
  if (!digits) {
    window.alert('Este cliente no tiene teléfono cargado.');
    return;
  }
  window.location.href = `tel:${digits}`;
}

function orderAddress(order) {
  return [order.client_address, order.locality_name, order.locality_province].filter(Boolean).join(', ');
}

function googleMapsRouteUrl(orders) {
  const stops = orders.map(orderAddress).filter(Boolean);
  if (!stops.length) return null;
  if (stops.length === 1) {
    return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(stops[0])}`;
  }
  const origin = encodeURIComponent(stops[0]);
  const destination = encodeURIComponent(stops[stops.length - 1]);
  const mid = stops.slice(1, -1).map(encodeURIComponent).join('|');
  let url = `https://www.google.com/maps/dir/?api=1&origin=${origin}&destination=${destination}&travelmode=driving`;
  if (mid) url += `&waypoints=${mid}`;
  return url;
}

function formatLoadDate(value) {
  if (!value) return '—';
  const s = String(value).slice(0, 10);
  const [y, m, d] = s.split('-');
  if (!y || !m || !d) return s;
  return `${d}/${m}/${y}`;
}

function printOrder(order) {
  const addr = orderAddress(order) || 'Sin dirección';
  const product = order.product_label || order.product_name || '';
  const html = `<!doctype html><html lang="es"><head><meta charset="utf-8"><title>Orden #${order.id}</title>
    <style>
      body{font-family:"Segoe UI",sans-serif;padding:24px;color:#1c1917}
      h1{margin:0 0 4px;font-size:22px}
      .muted{color:#6b645c}
      table{width:100%;border-collapse:collapse;margin-top:16px}
      th,td{border:1px solid #d6d3d1;padding:8px 10px;text-align:left}
      th{width:180px;background:#f5f5f4}
    </style></head><body>
    <h1>Orden de servicio #${order.id}</h1>
    <p class="muted">Fecha carga: ${formatLoadDate(order.load_date || order.client_created_at)} · ${order.status_name || ''}</p>
    <table>
      <tr><th>Estado</th><td>${order.status_name || ''}</td></tr>
      <tr><th>Fecha carga</th><td>${formatLoadDate(order.load_date || order.client_created_at)}</td></tr>
      <tr><th>Cliente</th><td>${order.client_name || ''}</td></tr>
      <tr><th>Teléfono</th><td>${order.client_phone || ''}</td></tr>
      <tr><th>Dirección</th><td>${addr}</td></tr>
      <tr><th>Localidad</th><td>${order.locality_name || ''}</td></tr>
      <tr><th>Tipo de producto</th><td>${order.product_type_name || ''}</td></tr>
      <tr><th>Producto</th><td>${product}</td></tr>
      <tr><th>Origen</th><td>${order.provider_name || 'Carga propia'}${order.provider_kind ? ` (${order.provider_kind})` : ''}</td></tr>
      <tr><th>Trabajo</th><td>${order.title || ''}</td></tr>
      <tr><th>Detalle</th><td>${order.description || ''}</td></tr>
    </table>
    <p class="muted">Instal Service S.A. — plantilla provisional. Podés enviar tu formato de orden para replicarlo.</p>
    <script>window.onload=()=>window.print()<\/script>
    </body></html>`;
  const w = window.open('', '_blank', 'noopener,noreferrer,width=800,height=900');
  if (!w) return;
  w.document.write(html);
  w.document.close();
}

function Field({ label, children }) {
  return (
    <label>
      {label}
      {children}
    </label>
  );
}

function Login({ onOk }) {
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('admin123');
  const [error, setError] = useState('');
  async function submit(e) {
    e.preventDefault();
    setError('');
    try {
      const data = await api('/api/auth/login', { method: 'POST', body: { username, password } });
      setToken(data.token);
      onOk(data.user);
    } catch (err) {
      setError(err.message);
    }
  }
  return (
    <div className="login">
      <form className="login-card" onSubmit={submit}>
        <img className="brand-logo" src="/logo-instal-service.jpg" alt="Instal Service S.A." />
        <h1>Instal Service S.A.</h1>
        <p>Gestión de taller · clima y línea blanca</p>
        <div className="grid">
          <Field label="Usuario">
            <input value={username} onChange={(e) => setUsername(e.target.value)} />
          </Field>
          <Field label="Clave">
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
          </Field>
          {error && <div className="error">{error}</div>}
          <button className="primary cta" type="submit">Entrar</button>
        </div>
        <p className="hint">Demo: admin/admin123 (Administrador) · diego/diego123 (Técnico)</p>
      </form>
    </div>
  );
}

function Filters({ lookups, filters, setFilters, extra }) {
  const set = (k) => (e) => setFilters((f) => ({ ...f, [k]: e.target.value }));
  return (
    <div className="row" style={{ marginBottom: 16 }}>
      <Field label="Fecha">
        <input type="date" value={filters.date || ''} onChange={set('date')} />
      </Field>
      <Field label="Técnico">
        <select value={filters.technicianId || ''} onChange={set('technicianId')}>
          <option value="">Todos</option>
          {lookups.technicians?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
        </select>
      </Field>
      <Field label="Tipo de producto">
        <select value={filters.productTypeId || ''} onChange={set('productTypeId')}>
          <option value="">Todos</option>
          {lookups.productTypes?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
        </select>
      </Field>
      <Field label="Proveedor / prestador">
        <select value={filters.providerId || ''} onChange={set('providerId')}>
          <option value="">Todos</option>
          {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name} ({p.kind})</option>)}
        </select>
      </Field>
      <Field label="Localidad">
        <select value={filters.localityId || ''} onChange={set('localityId')}>
          <option value="">Todas</option>
          {lookups.localities?.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
        </select>
      </Field>
      {extra}
    </div>
  );
}

function qs(filters) {
  const p = new URLSearchParams();
  Object.entries(filters).forEach(([k, v]) => { if (v) p.set(k, v); });
  const s = p.toString();
  return s ? `?${s}` : '';
}

function SheetInput({ value, onCommit, type = 'text', disabled }) {
  const [v, setV] = useState(value ?? '');
  useEffect(() => { setV(value ?? ''); }, [value]);
  return (
    <input
      className="sheet-input"
      type={type}
      value={v}
      disabled={disabled}
      onChange={(e) => setV(e.target.value)}
      onBlur={() => {
        if (String(v) !== String(value ?? '')) onCommit(v);
      }}
      onKeyDown={(e) => { if (e.key === 'Enter') e.currentTarget.blur(); }}
    />
  );
}

function SheetSelect({ value, onCommit, children, disabled }) {
  return (
    <select
      className="sheet-input"
      disabled={disabled}
      value={value ?? ''}
      onChange={(e) => onCommit(e.target.value)}
    >
      {children}
    </select>
  );
}

function Agenda({ lookups, user, onLookups }) {
  const canEditAll = user.role === 'admin' || user.role === 'coordinador';
  const isTech = user.role === 'tecnico';
  const [filters, setFilters] = useState({ date: today() });
  const [data, setData] = useState({ orders: [], technicians: [] });
  const [clients, setClients] = useState([]);
  const [newType, setNewType] = useState('');
  const [newLocality, setNewLocality] = useState('');
  const [draft, setDraft] = useState(null);
  const [msg, setMsg] = useState('');

  async function load() {
    setMsg('');
    try {
      const agenda = await api(`/api/agenda${qs(filters)}`);
      setData(agenda);
    } catch (err) {
      setMsg(err.message || 'No se pudo cargar la agenda');
      setData({ orders: [], technicians: [] });
    }
    try {
      setClients(await api('/api/clients'));
    } catch {
      /* la agenda igual se muestra */
    }
  }
  useEffect(() => { load().catch(console.error); }, [JSON.stringify(filters)]);

  async function patchOrder(order, fields) {
    try {
      const res = await api(`/api/orders/${order.id}`, { method: 'PATCH', body: fields });
      if (res.whatsapp?.auto_open && res.whatsapp.url && canEditAll) {
        if (window.confirm(`¿Abrir WhatsApp al cliente con la plantilla “${res.whatsapp.template}”?`)) {
          openWhatsApp(res.whatsapp.url);
        }
      }
      await load();
    } catch (err) {
      setMsg(err.message);
    }
  }

  async function sendWhatsApp(order) {
    try {
      const msg = await api(`/api/orders/${order.id}/whatsapp`);
      openWhatsApp(msg.url);
    } catch (err) {
      setMsg(err.message);
    }
  }

  async function patchClient(order, fields) {
    if (!order.client_id || !canEditAll) return;
    try {
      await api(`/api/clients/${order.client_id}`, { method: 'PATCH', body: fields });
      await load();
    } catch (err) {
      setMsg(err.message);
    }
  }

  async function addType(e) {
    e.preventDefault();
    if (!newType.trim()) return;
    await api('/api/product-types', { method: 'POST', body: { name: newType.trim() } });
    setNewType('');
    onLookups?.();
  }

  async function renameType(id, name) {
    if (!name.trim()) return;
    await api(`/api/product-types/${id}`, { method: 'PATCH', body: { name: name.trim() } });
    onLookups?.();
    load();
  }

  async function addLocality(e) {
    e.preventDefault();
    if (!newLocality.trim()) return;
    await api('/api/localities', { method: 'POST', body: { name: newLocality.trim() } });
    setNewLocality('');
    onLookups?.();
  }

  async function addRow(e) {
    e.preventDefault();
    if (!draft?.client_id) {
      setMsg('La fila nueva necesita un cliente.');
      return;
    }
    const productLabel = (draft.product_label || '').trim();
    await api('/api/orders', {
      method: 'POST',
      body: {
        client_id: Number(draft.client_id),
        technician_id: draft.technician_id ? Number(draft.technician_id) : null,
        product_type_id: draft.product_type_id ? Number(draft.product_type_id) : null,
        product_label: productLabel || null,
        provider_id: draft.provider_id ? Number(draft.provider_id) : null,
        locality_id: draft.locality_id ? Number(draft.locality_id) : null,
        status_id: Number(draft.status_id),
        title: productLabel || 'Servicio',
        scheduled_date: filters.date || today(),
      },
    });
    setDraft(null);
    setMsg('');
    load();
  }

  function startDraft() {
    setDraft({
      client_id: '',
      technician_id: '',
      product_type_id: '',
      product_label: '',
      provider_id: '',
      locality_id: '',
      status_id: lookups.statuses?.[0]?.id || '',
    });
  }

  const rows = data.orders || [];
  const route = googleMapsRouteUrl(rows);

  return (
    <section>
      <h2>{isTech ? 'Mi agenda del día' : 'Agenda diaria'}</h2>
      <p className="meta">
        Vista en lista, como una planilla: cada celda se edita al escribir y se guarda al salir del campo.
        {canEditAll ? ' Como administración podés cambiar todos los campos y dar de alta tipos de producto.' : ' Solo tus órdenes del día.'}
      </p>
      {isTech ? (
        <div className="row" style={{ marginBottom: 12 }}>
          <Field label="Fecha">
            <input type="date" value={filters.date || ''} onChange={(e) => setFilters({ date: e.target.value })} />
          </Field>
        </div>
      ) : (
        <Filters lookups={lookups} filters={filters} setFilters={setFilters} />
      )}

      <div className="row" style={{ marginBottom: 12 }}>
        <button className="ghost" type="button" disabled={!route} onClick={() => window.open(route, '_blank', 'noopener,noreferrer')}>
          Hoja de ruta (Maps)
        </button>
        {canEditAll && (
          <button className="primary" type="button" onClick={startDraft}>Nueva fila</button>
        )}
      </div>

      {canEditAll && (
        <div className="card" style={{ marginBottom: 12 }}>
          <strong>Tipos de producto</strong>
          <div className="type-chips" style={{ marginTop: 8 }}>
            {lookups.productTypes?.map((t) => (
              <SheetInput key={t.id} value={t.name} onCommit={(name) => renameType(t.id, name)} />
            ))}
            <form className="row" onSubmit={addType}>
              <input placeholder="Nuevo tipo (ej. Lavarropas)" value={newType} onChange={(e) => setNewType(e.target.value)} />
              <button className="ghost" type="submit">Agregar tipo</button>
            </form>
          </div>
          <form className="row" onSubmit={addLocality} style={{ marginTop: 8 }}>
            <input placeholder="Nueva localidad" value={newLocality} onChange={(e) => setNewLocality(e.target.value)} />
            <button className="ghost" type="submit">Agregar localidad</button>
          </form>
        </div>
      )}

      {msg && <p className="error">{msg}</p>}

      <div className="sheet-wrap">
        <table className="sheet">
          <thead>
            <tr>
              <th>Estado</th>
              <th>Fecha carga</th>
              <th>Cliente</th>
              <th>Teléfono</th>
              <th>Dirección</th>
              <th>Localidad</th>
              <th>Tipo de producto</th>
              <th>Producto</th>
              <th>Origen</th>
              <th>Imprimir</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((o) => (
              <tr key={o.id}>
                <td>
                  <SheetSelect value={o.status_id} onCommit={(v) => patchOrder(o, { status_id: Number(v) })}>
                    {lookups.statuses?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                  </SheetSelect>
                </td>
                <td className="sheet-readonly" title={o.client_source === 'excel' ? 'Alta por importación Excel' : 'Alta manual'}>
                  {formatLoadDate(o.load_date || o.client_created_at)}
                  <div className="meta" style={{ padding: '0 8px 4px' }}>
                    {o.client_source === 'excel' ? 'Importación' : 'Alta manual'}
                  </div>
                </td>
                <td>
                  <SheetSelect
                    value={o.client_id}
                    disabled={!canEditAll}
                    onCommit={(v) => patchOrder(o, { client_id: Number(v) })}
                  >
                    {clients.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                  </SheetSelect>
                </td>
                <td>
                  <SheetInput value={o.client_phone || ''} disabled={!canEditAll} onCommit={(v) => patchClient(o, { phone: v })} />
                </td>
                <td>
                  <SheetInput value={o.client_address || ''} disabled={!canEditAll} onCommit={(v) => patchClient(o, { address: v })} />
                </td>
                <td>
                  <SheetSelect value={o.locality_id || ''} onCommit={(v) => patchOrder(o, { locality_id: v ? Number(v) : null })}>
                    <option value="">—</option>
                    {lookups.localities?.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
                  </SheetSelect>
                </td>
                <td>
                  <SheetSelect value={o.product_type_id || ''} onCommit={(v) => patchOrder(o, { product_type_id: v ? Number(v) : null })}>
                    <option value="">—</option>
                    {lookups.productTypes?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                  </SheetSelect>
                </td>
                <td>
                  <SheetInput
                    value={o.product_label || ''}
                    onCommit={(v) => patchOrder(o, { product_label: v, title: v || o.title })}
                  />
                </td>
                <td>
                  <SheetSelect
                    value={o.provider_id || ''}
                    disabled={!canEditAll}
                    onCommit={(v) => patchOrder(o, { provider_id: v ? Number(v) : null })}
                  >
                    <option value="">Carga propia</option>
                    {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name} ({p.kind})</option>)}
                  </SheetSelect>
                </td>
                <td className="sheet-actions">
                  {isTech && (
                    <button className="primary sheet-btn" type="button" onClick={() => callClient(o.client_phone)}>
                      Llamar
                    </button>
                  )}
                  {(canEditAll || isTech) && (
                    <button className="ghost sheet-btn wa-btn" type="button" onClick={() => sendWhatsApp(o)}>
                      WhatsApp
                    </button>
                  )}
                  <button className="ghost sheet-btn" type="button" onClick={() => printOrder(o)}>Imprimir</button>
                </td>
              </tr>
            ))}
            {draft && (
              <tr className="sheet-draft">
                <td>
                  <SheetSelect value={draft.status_id} onCommit={(v) => setDraft({ ...draft, status_id: v })}>
                    {lookups.statuses?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                  </SheetSelect>
                </td>
                <td className="sheet-readonly meta" style={{ padding: 8 }}>Al guardar</td>
                <td>
                  <SheetSelect value={draft.client_id} onCommit={(v) => setDraft({ ...draft, client_id: v })}>
                    <option value="">Cliente…</option>
                    {clients.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                  </SheetSelect>
                </td>
                <td colSpan={2} className="meta" style={{ padding: 8 }}>Se completa con el cliente</td>
                <td>
                  <SheetSelect value={draft.locality_id} onCommit={(v) => setDraft({ ...draft, locality_id: v })}>
                    <option value="">—</option>
                    {lookups.localities?.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
                  </SheetSelect>
                </td>
                <td>
                  <SheetSelect value={draft.product_type_id} onCommit={(v) => setDraft({ ...draft, product_type_id: v })}>
                    <option value="">—</option>
                    {lookups.productTypes?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                  </SheetSelect>
                </td>
                <td>
                  <SheetInput value={draft.product_label} onCommit={(v) => setDraft({ ...draft, product_label: v })} />
                </td>
                <td>
                  <SheetSelect value={draft.provider_id} onCommit={(v) => setDraft({ ...draft, provider_id: v })}>
                    <option value="">Carga propia</option>
                    {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
                  </SheetSelect>
                </td>
                <td>
                  <button className="primary sheet-btn" type="button" onClick={addRow}>Guardar fila</button>
                </td>
              </tr>
            )}
            {rows.length === 0 && !draft && (
              <tr><td colSpan={10} className="meta" style={{ padding: 12 }}>No hay órdenes en esta fecha. {canEditAll && 'Usá “Nueva fila” para cargar una.'}</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function Seguimientos() {
  const [items, setItems] = useState([]);
  const [status, setStatus] = useState('pendiente');
  async function load() {
    setItems(await api(`/api/followups?status=${status}`));
  }
  useEffect(() => { load().catch(console.error); }, [status]);
  async function done(id) {
    await api(`/api/followups/${id}`, { method: 'PATCH', body: { status: 'hecho' } });
    load();
  }
  return (
    <section>
      <h2>Seguimientos</h2>
      <p className="meta">Se generan solos cuando una orden entra a un estado o supera el tiempo definido.</p>
      <select value={status} onChange={(e) => setStatus(e.target.value)} style={{ margin: '8px 0 16px' }}>
        <option value="pendiente">Pendientes</option>
        <option value="hecho">Hechos</option>
        <option value="todos">Todos</option>
      </select>
      <div className="cards">
        {items.map((f) => (
          <article className="card" key={f.id}>
            <h3>{f.rule_name || 'Seguimiento'}</h3>
            <p>{f.order_title} · {f.client_name}</p>
            <p className="meta">Estado orden: {f.order_status} · Asignado: {f.assigned_name || '—'}</p>
            {f.status === 'pendiente' && (
              <button className="primary" onClick={() => done(f.id)}>Marcar hecho</button>
            )}
          </article>
        ))}
        {items.length === 0 && <p className="meta">No hay seguimientos en esta vista.</p>}
      </div>
    </section>
  );
}

const MONTHS = [
  { value: '', label: 'Todos' },
  { value: '01', label: 'Enero' },
  { value: '02', label: 'Febrero' },
  { value: '03', label: 'Marzo' },
  { value: '04', label: 'Abril' },
  { value: '05', label: 'Mayo' },
  { value: '06', label: 'Junio' },
  { value: '07', label: 'Julio' },
  { value: '08', label: 'Agosto' },
  { value: '09', label: 'Septiembre' },
  { value: '10', label: 'Octubre' },
  { value: '11', label: 'Noviembre' },
  { value: '12', label: 'Diciembre' },
];

function monthLabel(mm) {
  return MONTHS.find((m) => m.value === mm)?.label || mm;
}

function Ordenes({ lookups, onCreated }) {
  const currentYear = String(new Date().getFullYear());
  const [filters, setFilters] = useState({ year: currentYear, month: '' });
  const [years, setYears] = useState([currentYear]);
  const [rows, setRows] = useState([]);
  const [total, setTotal] = useState(0);
  const [clients, setClients] = useState([]);
  const [openMonths, setOpenMonths] = useState({});
  const [form, setForm] = useState({
    title: '', client_id: '', technician_id: '', product_type_id: '', product_label: '', provider_id: '',
    locality_id: '', status_id: lookups.statuses?.[0]?.id || '', scheduled_date: today(),
    description: '',
  });

  async function load() {
    const data = await api(`/api/orders${qs(filters)}`);
    const list = Array.isArray(data) ? data : (data.orders || []);
    setRows(list);
    setTotal(data.total ?? list.length);
    const y = data.years?.length ? data.years : [currentYear];
    setYears(y);
    if (filters.year && !y.includes(filters.year) && y[0]) {
      setFilters((f) => ({ ...f, year: y[0] }));
    }
    // Abrir todos los meses con datos cuando el filtro es "Todos"
    if (!filters.month) {
      const open = {};
      for (const o of list) {
        if (o.entry_month) open[o.entry_month] = true;
      }
      setOpenMonths(open);
    }
    setClients(await api('/api/clients'));
  }
  useEffect(() => { load().catch(console.error); }, [JSON.stringify(filters)]);
  useEffect(() => {
    if (lookups.statuses?.[0] && !form.status_id) {
      setForm((f) => ({ ...f, status_id: lookups.statuses[0].id }));
    }
  }, [lookups.statuses]);

  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));
  const setFilter = (k) => (e) => setFilters((f) => ({ ...f, [k]: e.target.value }));

  async function create(e) {
    e.preventDefault();
    const productLabel = (form.product_label || '').trim();
    await api('/api/orders', {
      method: 'POST',
      body: {
        client_id: Number(form.client_id),
        technician_id: form.technician_id ? Number(form.technician_id) : null,
        product_type_id: form.product_type_id ? Number(form.product_type_id) : null,
        product_label: productLabel || null,
        provider_id: form.provider_id ? Number(form.provider_id) : null,
        locality_id: form.locality_id ? Number(form.locality_id) : null,
        status_id: Number(form.status_id),
        title: form.title || productLabel || 'Servicio',
        description: form.description || null,
        scheduled_date: form.scheduled_date || today(),
      },
    });
    onCreated?.();
    load();
  }

  const byMonth = {};
  for (const o of rows) {
    const key = o.entry_month || '00';
    if (!byMonth[key]) byMonth[key] = [];
    byMonth[key].push(o);
  }
  const monthKeys = Object.keys(byMonth).sort((a, b) => b.localeCompare(a));

  function renderTable(list) {
    return (
      <table>
        <thead>
          <tr>
            <th>Fecha ingreso</th>
            <th>Cliente</th>
            <th>Producto</th>
            <th>Tipo</th>
            <th>Origen</th>
            <th>Estado</th>
            <th>Localidad</th>
          </tr>
        </thead>
        <tbody>
          {list.map((o) => (
            <tr key={o.id}>
              <td>
                <strong>{formatLoadDate(o.entry_date || o.created_at)}</strong>
                <div className="meta">#{o.id}</div>
              </td>
              <td>{o.client_name}</td>
              <td>{o.product_label || o.title || '—'}</td>
              <td>{o.product_type_name || '—'}</td>
              <td>{o.provider_name || 'Carga propia'}</td>
              <td><span className="badge" style={{ background: o.status_color }}>{o.status_name}</span></td>
              <td>{o.locality_name || '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }

  return (
    <section>
      <h2>Histórico de órdenes</h2>
      <p className="meta">
        Base histórica de todos los ingresos (vigentes y cumplidos). El mismo cliente puede reingresar;
        cada visita se diferencia por su <strong>fecha de ingreso</strong>.
      </p>

      <div className="history-filters card row" style={{ marginBottom: 16 }}>
        <Field label="Año">
          <select value={filters.year || ''} onChange={setFilter('year')}>
            {years.map((y) => <option key={y} value={y}>{y}</option>)}
          </select>
        </Field>
        <Field label="Mes">
          <select value={filters.month || ''} onChange={setFilter('month')}>
            {MONTHS.map((m) => <option key={m.value || 'all'} value={m.value}>{m.label}</option>)}
          </select>
        </Field>
        <Field label="Estado">
          <select value={filters.statusId || ''} onChange={setFilter('statusId')}>
            <option value="">Todos</option>
            {lookups.statuses?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>
        </Field>
        <Field label="Origen">
          <select value={filters.providerId || ''} onChange={setFilter('providerId')}>
            <option value="">Todos</option>
            {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select>
        </Field>
        <Field label="Tipo de producto">
          <select value={filters.productTypeId || ''} onChange={setFilter('productTypeId')}>
            <option value="">Todos</option>
            {lookups.productTypes?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
          </select>
        </Field>
        <p className="meta" style={{ alignSelf: 'center' }}>{total} registro{total === 1 ? '' : 's'}</p>
      </div>

      <form className="card grid" onSubmit={create} style={{ marginBottom: 16 }}>
        <strong>Nuevo ingreso / reingreso</strong>
        <p className="meta">Si el cliente ya existía, se crea otra orden con fecha de ingreso de hoy.</p>
        <div className="row">
          <Field label="Cliente">
            <select required value={form.client_id} onChange={set('client_id')}>
              <option value="">Elegir</option>
              {clients.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </Field>
          <Field label="Producto (texto libre)"><input value={form.product_label} onChange={set('product_label')} placeholder="Ej. Lavarropas Drean 8kg" /></Field>
          <Field label="Tipo">
            <select value={form.product_type_id} onChange={set('product_type_id')}>
              <option value="">—</option>
              {lookups.productTypes?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
            </select>
          </Field>
          <Field label="Origen">
            <select value={form.provider_id} onChange={set('provider_id')}>
              <option value="">Carga propia</option>
              {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select>
          </Field>
          <Field label="Localidad">
            <select value={form.locality_id} onChange={set('locality_id')}>
              <option value="">—</option>
              {lookups.localities?.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
            </select>
          </Field>
          <Field label="Estado">
            <select value={form.status_id} onChange={set('status_id')}>
              {lookups.statuses?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
          </Field>
          <Field label="Técnico">
            <select value={form.technician_id} onChange={set('technician_id')}>
              <option value="">Sin asignar</option>
              {lookups.technicians?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
            </select>
          </Field>
          <Field label="Agenda (opcional)"><input type="date" value={form.scheduled_date} onChange={set('scheduled_date')} /></Field>
        </div>
        <Field label="Detalle"><textarea value={form.description} onChange={set('description')} /></Field>
        <button className="primary" type="submit">Registrar ingreso</button>
      </form>

      {!filters.month && monthKeys.length > 0 && (
        <div className="month-panels">
          {monthKeys.map((mm) => (
            <details
              key={mm}
              className="month-panel"
              open={openMonths[mm] !== false}
              onToggle={(e) => setOpenMonths((prev) => ({ ...prev, [mm]: e.target.open }))}
            >
              <summary>
                <span>{monthLabel(mm)} {filters.year}</span>
                <span className="meta">{byMonth[mm].length} ingreso{byMonth[mm].length === 1 ? '' : 's'}</span>
              </summary>
              {renderTable(byMonth[mm])}
            </details>
          ))}
        </div>
      )}

      {filters.month && (
        <div className="card">
          <h3 style={{ marginTop: 0 }}>{monthLabel(filters.month)} {filters.year}</h3>
          {rows.length ? renderTable(rows) : <p className="meta">Sin ingresos en este mes.</p>}
        </div>
      )}

      {!rows.length && (
        <p className="meta">No hay órdenes para este período.</p>
      )}
    </section>
  );
}

function money(n) {
  return Number(n || 0).toLocaleString('es-AR', { style: 'currency', currency: 'ARS', maximumFractionDigits: 0 });
}

function Clientes({ lookups, onCreated }) {
  const [rows, setRows] = useState([]);
  const [view, setView] = useState('all');
  const [form, setForm] = useState({ name: '', phone: '', email: '', provider_id: '', locality_id: '', external_id: '', address: '', notes: '' });
  const [importProvider, setImportProvider] = useState('');
  const [importMsg, setImportMsg] = useState('');
  const [busy, setBusy] = useState(false);

  function query() {
    if (view === 'all') return '';
    if (view === 'manual') return '?source=manual';
    if (view === 'excel') return '?source=excel';
    if (view === 'own') return '?providerId=own';
    return `?providerId=${view}`;
  }

  async function load() {
    setRows(await api(`/api/clients${query()}`));
  }
  useEffect(() => { load().catch(console.error); }, [view]);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));

  async function create(e) {
    e.preventDefault();
    await api('/api/clients', {
      method: 'POST',
      body: {
        ...form,
        provider_id: form.provider_id ? Number(form.provider_id) : null,
        locality_id: form.locality_id ? Number(form.locality_id) : null,
      },
    });
    setForm({ name: '', phone: '', email: '', provider_id: form.provider_id, locality_id: '', external_id: '', address: '', notes: '' });
    onCreated?.();
    load();
  }

  async function importExcel(e) {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file || !importProvider) {
      setImportMsg('Elegí prestador/proveedor y un Excel.');
      return;
    }
    setBusy(true);
    setImportMsg('');
    try {
      const body = new FormData();
      body.append('file', file);
      const result = await api(`/api/clients/import?providerId=${importProvider}`, { method: 'POST', body });
      setImportMsg(`Importados: ${result.created} nuevos, ${result.updated} actualizados.${result.errors?.length ? ` Avisos: ${result.errors.slice(0, 3).join(' · ')}` : ''}`);
      setView(String(importProvider));
      onCreated?.();
    } catch (err) {
      setImportMsg(err.message);
    } finally {
      setBusy(false);
    }
  }

  const originLabel = (c) => {
    if (!c.provider_name) return 'Carga propia';
    return `${c.provider_name} (${c.provider_kind})`;
  };

  return (
    <section>
      <h2>Clientes (usuarios finales)</h2>
      <p className="meta">Vista unificada del servicio. Cada ficha queda marcada si nació por carga manual o por Excel de un prestador/proveedor (Salesforce).</p>
      <div className="tabs">
        <button type="button" className={view === 'all' ? 'active' : ''} onClick={() => setView('all')}>Unificada</button>
        <button type="button" className={view === 'manual' ? 'active' : ''} onClick={() => setView('manual')}>Carga manual</button>
        <button type="button" className={view === 'excel' ? 'active' : ''} onClick={() => setView('excel')}>Importados Excel</button>
        <button type="button" className={view === 'own' ? 'active' : ''} onClick={() => setView('own')}>Sin prestador</button>
        {lookups.providers?.map((p) => (
          <button type="button" key={p.id} className={view === String(p.id) ? 'active' : ''} onClick={() => setView(String(p.id))}>
            {p.name}
          </button>
        ))}
      </div>

      <form className="card grid" onSubmit={create} style={{ margin: '16px 0' }}>
        <strong>Alta manual</strong>
        <div className="row">
          <Field label="Nombre"><input required value={form.name} onChange={set('name')} /></Field>
          <Field label="Teléfono"><input value={form.phone} onChange={set('phone')} /></Field>
          <Field label="Email"><input value={form.email} onChange={set('email')} /></Field>
          <Field label="Prestador / proveedor">
            <select value={form.provider_id} onChange={set('provider_id')}>
              <option value="">Carga propia</option>
              {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name} ({p.kind})</option>)}
            </select>
          </Field>
          <Field label="ID Salesforce"><input value={form.external_id} onChange={set('external_id')} /></Field>
          <Field label="Localidad">
            <select value={form.locality_id} onChange={set('locality_id')}>
              <option value="">—</option>
              {lookups.localities?.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
            </select>
          </Field>
          <Field label="Dirección"><input value={form.address} onChange={set('address')} /></Field>
          <Field label="Notas"><input value={form.notes} onChange={set('notes')} /></Field>
        </div>
        <button className="primary" type="submit">Crear cliente</button>
      </form>

      <div className="card grid" style={{ marginBottom: 16 }}>
        <strong>Importar Excel de Salesforce</strong>
        <p className="meta">Pedí a tus prestadores/proveedores que exporten con estas columnas (o equivalentes en inglés): ID_Salesforce, Nombre, Telefono, Email, Direccion, Localidad, Provincia, Notas. Si el ID ya existe para ese origen, se actualiza la ficha.</p>
        <div className="row">
          <Field label="Origen de este archivo">
            <select value={importProvider} onChange={(e) => setImportProvider(e.target.value)}>
              <option value="">Elegir prestador/proveedor</option>
              {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name} ({p.kind})</option>)}
            </select>
          </Field>
          <button className="ghost" type="button" onClick={() => downloadFile('/api/clients/template.xlsx', 'plantilla-clientes-salesforce.xlsx')}>
            Descargar plantilla
          </button>
          <Field label="Archivo .xlsx">
            <input type="file" accept=".xlsx,.xls" disabled={busy} onChange={importExcel} />
          </Field>
        </div>
        {importMsg && <p className="meta">{importMsg}</p>}
      </div>

      <table>
        <thead>
          <tr>
            <th>Cliente</th><th>Prestador / proveedor</th><th>Alta</th><th>ID Salesforce</th><th>Localidad</th><th>Teléfono</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((c) => (
            <tr key={c.id}>
              <td>{c.name}</td>
              <td>{originLabel(c)}</td>
              <td>{c.source === 'excel' ? 'Excel' : 'Manual'}</td>
              <td>{c.external_id || '—'}</td>
              <td>{c.locality_name || '—'}</td>
              <td>{c.phone}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Tarifas({ lookups }) {
  const [rows, setRows] = useState([]);
  const [form, setForm] = useState({
    name: '', scope: 'proveedor', provider_id: '', technician_id: '',
    income_fixed: 0, income_per_hour: 0, income_per_km: 0, income_parts_pct: 0,
    cost_fixed: 0, cost_per_hour: 0, cost_per_km: 0, cost_parts_pct: 100,
  });
  async function load() { setRows(await api('/api/tariffs')); }
  useEffect(() => { load().catch(console.error); }, []);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));
  async function create(e) {
    e.preventDefault();
    await api('/api/tariffs', {
      method: 'POST',
      body: {
        ...form,
        provider_id: form.scope === 'proveedor' ? Number(form.provider_id) : null,
        technician_id: form.scope === 'tecnico' ? Number(form.technician_id) : null,
      },
    });
    load();
  }
  return (
    <section>
      <h2>Tarifario</h2>
      <p className="meta">
        Fijo: monto por orden. Variable: horas, km y % sobre venta de repuestos (ingreso) o sobre costo de repuestos (egreso).
        La tarifa de prestador/proveedor calcula lo que cobrás; la de técnico, lo que te cuesta.
      </p>
      <form className="card grid" onSubmit={create} style={{ marginBottom: 16 }}>
        <div className="row">
          <Field label="Nombre"><input required value={form.name} onChange={set('name')} /></Field>
          <Field label="Aplica a">
            <select value={form.scope} onChange={set('scope')}>
              <option value="proveedor">Prestador / proveedor (ingreso)</option>
              <option value="tecnico">Técnico (costo)</option>
              <option value="general">General (respaldo)</option>
            </select>
          </Field>
          {form.scope === 'proveedor' && (
            <Field label="Prestador / proveedor">
              <select required value={form.provider_id} onChange={set('provider_id')}>
                <option value="">Elegir</option>
                {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
              </select>
            </Field>
          )}
          {form.scope === 'tecnico' && (
            <Field label="Técnico">
              <select required value={form.technician_id} onChange={set('technician_id')}>
                <option value="">Elegir</option>
                {lookups.technicians?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </Field>
          )}
        </div>
        <strong>Parte fija y variable de ingreso</strong>
        <div className="row">
          <Field label="Fijo por orden"><input type="number" value={form.income_fixed} onChange={set('income_fixed')} /></Field>
          <Field label="$ por hora"><input type="number" value={form.income_per_hour} onChange={set('income_per_hour')} /></Field>
          <Field label="$ por km"><input type="number" value={form.income_per_km} onChange={set('income_per_km')} /></Field>
          <Field label="% sobre venta de repuestos"><input type="number" value={form.income_parts_pct} onChange={set('income_parts_pct')} /></Field>
        </div>
        <strong>Parte fija y variable de costo</strong>
        <div className="row">
          <Field label="Fijo por orden"><input type="number" value={form.cost_fixed} onChange={set('cost_fixed')} /></Field>
          <Field label="$ por hora"><input type="number" value={form.cost_per_hour} onChange={set('cost_per_hour')} /></Field>
          <Field label="$ por km"><input type="number" value={form.cost_per_km} onChange={set('cost_per_km')} /></Field>
          <Field label="% del costo de repuestos"><input type="number" value={form.cost_parts_pct} onChange={set('cost_parts_pct')} /></Field>
        </div>
        <button className="primary" type="submit">Guardar tarifa</button>
      </form>
      <table>
        <thead>
          <tr>
            <th>Tarifa</th><th>Alcance</th><th>Ingreso fijo</th><th>$/h</th><th>$/km</th><th>% rep.</th><th>Costo fijo</th><th>Costo/h</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((t) => (
            <tr key={t.id}>
              <td>{t.name}<div className="meta">{t.provider_name || t.technician_name || 'General'}</div></td>
              <td>{t.scope}</td>
              <td>{money(t.income_fixed)}</td>
              <td>{money(t.income_per_hour)}</td>
              <td>{money(t.income_per_km)}</td>
              <td>{t.income_parts_pct}%</td>
              <td>{money(t.cost_fixed)}</td>
              <td>{money(t.cost_per_hour)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Resultados() {
  const [group, setGroup] = useState('proveedor');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const [data, setData] = useState({ rows: [], totals: { orders: 0, income: 0, cost: 0, profit: 0 } });
  async function load() {
    const q = new URLSearchParams({ group });
    if (from) q.set('from', from);
    if (to) q.set('to', to);
    setData(await api(`/api/results?${q}`));
  }
  useEffect(() => { load().catch(console.error); }, [group, from, to]);
  return (
    <section>
      <h2>Ganancias y costos</h2>
      <p className="meta">Se calcula con el tarifario (fijo + horas/km/repuestos) de cada orden. Podés ver el resultado por prestador/proveedor, por cliente o por técnico.</p>
      <div className="row" style={{ marginBottom: 16 }}>
        <Field label="Agrupar">
          <select value={group} onChange={(e) => setGroup(e.target.value)}>
            <option value="proveedor">Prestador / proveedor</option>
            <option value="cliente">Cliente</option>
            <option value="tecnico">Técnico</option>
          </select>
        </Field>
        <Field label="Desde"><input type="date" value={from} onChange={(e) => setFrom(e.target.value)} /></Field>
        <Field label="Hasta"><input type="date" value={to} onChange={(e) => setTo(e.target.value)} /></Field>
      </div>
      <div className="cards" style={{ marginBottom: 16 }}>
        <article className="card"><h3>Órdenes</h3><p>{data.totals.orders}</p></article>
        <article className="card"><h3>Ingresos</h3><p>{money(data.totals.income)}</p></article>
        <article className="card"><h3>Costos</h3><p>{money(data.totals.cost)}</p></article>
        <article className="card"><h3>Resultado</h3><p>{money(data.totals.profit)}</p></article>
      </div>
      <table>
        <thead>
          <tr>
            <th>{group === 'cliente' ? 'Cliente' : group === 'tecnico' ? 'Técnico' : 'Prestador / proveedor'}</th>
            <th>Órdenes</th><th>Ingresos</th><th>Costos</th><th>Ganancia</th>
          </tr>
        </thead>
        <tbody>
          {data.rows.map((r) => (
            <tr key={r.key}>
              <td>{r.label}</td>
              <td>{r.orders}</td>
              <td>{money(r.income)}</td>
              <td>{money(r.cost)}</td>
              <td>{money(r.profit)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Catalogo({ lookups, onCreated }) {
  const [rows, setRows] = useState([]);
  const [form, setForm] = useState({ name: '', sku: '', product_type_id: '', provider_id: '', stock: 0, price: 0 });
  const [typeName, setTypeName] = useState('');
  async function load() { setRows(await api('/api/products')); }
  useEffect(() => { load().catch(console.error); }, []);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));
  async function create(e) {
    e.preventDefault();
    await api('/api/products', {
      method: 'POST',
      body: {
        ...form,
        stock: Number(form.stock),
        price: Number(form.price),
        product_type_id: form.product_type_id ? Number(form.product_type_id) : null,
        provider_id: form.provider_id ? Number(form.provider_id) : null,
      },
    });
    load();
  }
  async function addType(e) {
    e.preventDefault();
    await api('/api/product-types', { method: 'POST', body: { name: typeName } });
    setTypeName('');
    onCreated?.();
  }
  return (
    <section>
      <h2>Repuestos</h2>
      <form className="row" onSubmit={addType} style={{ marginBottom: 12 }}>
        <Field label="Nuevo tipo"><input value={typeName} onChange={(e) => setTypeName(e.target.value)} /></Field>
        <button className="ghost" type="submit">Agregar tipo</button>
      </form>
      <form className="card row" onSubmit={create} style={{ marginBottom: 16 }}>
        <Field label="Nombre"><input required value={form.name} onChange={set('name')} /></Field>
        <Field label="SKU"><input value={form.sku} onChange={set('sku')} /></Field>
        <Field label="Tipo">
          <select value={form.product_type_id} onChange={set('product_type_id')}>
            <option value="">—</option>
            {lookups.productTypes?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
          </select>
        </Field>
        <Field label="Proveedor">
          <select value={form.provider_id} onChange={set('provider_id')}>
            <option value="">—</option>
            {lookups.providers?.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select>
        </Field>
        <Field label="Stock"><input type="number" value={form.stock} onChange={set('stock')} /></Field>
        <Field label="Precio"><input type="number" value={form.price} onChange={set('price')} /></Field>
        <button className="primary" type="submit">Alta</button>
      </form>
      <table>
        <thead><tr><th>Repuesto</th><th>Tipo</th><th>Proveedor</th><th>Stock</th><th>Precio</th></tr></thead>
        <tbody>
          {rows.map((p) => (
            <tr key={p.id}>
              <td>{p.name} <span className="meta">{p.sku}</span></td>
              <td>{p.type_name}</td>
              <td>{p.provider_name}</td>
              <td>{p.stock}</td>
              <td>${Number(p.price).toLocaleString('es-AR')}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Tecnicos({ onCreated }) {
  const [rows, setRows] = useState([]);
  const [form, setForm] = useState({ name: '', phone: '', specialty: '' });
  async function load() { setRows(await api('/api/technicians')); }
  useEffect(() => { load().catch(console.error); }, []);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));
  async function create(e) {
    e.preventDefault();
    await api('/api/technicians', { method: 'POST', body: form });
    onCreated?.();
    load();
  }
  return (
    <section>
      <h2>Técnicos / mecánicos</h2>
      <form className="card row" onSubmit={create} style={{ marginBottom: 16 }}>
        <Field label="Nombre"><input required value={form.name} onChange={set('name')} /></Field>
        <Field label="Teléfono"><input value={form.phone} onChange={set('phone')} /></Field>
        <Field label="Especialidad"><input value={form.specialty} onChange={set('specialty')} /></Field>
        <button className="primary" type="submit">Alta</button>
      </form>
      <table>
        <thead><tr><th>Nombre</th><th>Especialidad</th><th>Teléfono</th></tr></thead>
        <tbody>
          {rows.map((t) => (
            <tr key={t.id}><td>{t.name}</td><td>{t.specialty}</td><td>{t.phone}</td></tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Proveedores({ onCreated }) {
  const [rows, setRows] = useState([]);
  const [form, setForm] = useState({ name: '', kind: 'proveedor', contact: '', notes: '' });
  async function load() { setRows(await api('/api/providers')); }
  useEffect(() => { load().catch(console.error); }, []);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));
  async function create(e) {
    e.preventDefault();
    await api('/api/providers', { method: 'POST', body: form });
    onCreated?.();
    load();
  }
  return (
    <section>
      <h2>Proveedores y prestadores</h2>
      <p className="meta">Cada origen es una base distinta de clientes y/o repuestos.</p>
      <form className="card row" onSubmit={create} style={{ marginBottom: 16 }}>
        <Field label="Nombre"><input required value={form.name} onChange={set('name')} /></Field>
        <Field label="Tipo">
          <select value={form.kind} onChange={set('kind')}>
            <option value="proveedor">Proveedor</option>
            <option value="prestador">Prestador</option>
          </select>
        </Field>
        <Field label="Contacto"><input value={form.contact} onChange={set('contact')} /></Field>
        <Field label="Notas"><input value={form.notes} onChange={set('notes')} /></Field>
        <button className="primary" type="submit">Alta</button>
      </form>
      <table>
        <thead><tr><th>Nombre</th><th>Tipo</th><th>Contacto</th><th>Notas</th></tr></thead>
        <tbody>
          {rows.map((p) => (
            <tr key={p.id}><td>{p.name}</td><td>{p.kind}</td><td>{p.contact}</td><td>{p.notes}</td></tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Reglas({ lookups }) {
  const [rows, setRows] = useState([]);
  const [form, setForm] = useState({ name: '', trigger_type: 'status', status_id: '', hours: 48, assign_role: 'coordinador' });
  async function load() { setRows(await api('/api/followup-rules')); }
  useEffect(() => { load().catch(console.error); }, []);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));
  async function create(e) {
    e.preventDefault();
    await api('/api/followup-rules', {
      method: 'POST',
      body: {
        ...form,
        status_id: form.status_id ? Number(form.status_id) : null,
        hours: form.trigger_type === 'elapsed_hours' ? Number(form.hours) : null,
      },
    });
    load();
  }
  return (
    <section>
      <h2>Reglas de seguimiento</h2>
      <p className="meta">Por estado (ej. “Listo para entregar”) o por tiempo en un estado (ej. 48 h en diagnóstico).</p>
      <form className="card row" onSubmit={create} style={{ marginBottom: 16 }}>
        <Field label="Nombre"><input required value={form.name} onChange={set('name')} /></Field>
        <Field label="Disparo">
          <select value={form.trigger_type} onChange={set('trigger_type')}>
            <option value="status">Al entrar a un estado</option>
            <option value="elapsed_hours">Por tiempo en un estado</option>
          </select>
        </Field>
        <Field label="Estado">
          <select value={form.status_id} onChange={set('status_id')}>
            <option value="">—</option>
            {lookups.statuses?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>
        </Field>
        {form.trigger_type === 'elapsed_hours' && (
          <Field label="Horas"><input type="number" value={form.hours} onChange={set('hours')} /></Field>
        )}
        <Field label="Asignar a rol">
          <select value={form.assign_role} onChange={set('assign_role')}>
            <option value="coordinador">Coordinador</option>
            <option value="admin">Admin</option>
            <option value="tecnico">Técnico</option>
          </select>
        </Field>
        <button className="primary" type="submit">Crear regla</button>
      </form>
      <table>
        <thead><tr><th>Regla</th><th>Tipo</th><th>Estado</th><th>Horas</th><th>Rol</th></tr></thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.id}>
              <td>{r.name}</td>
              <td>{r.trigger_type === 'status' ? 'Por estado' : 'Por tiempo'}</td>
              <td>{r.status_name || '—'}</td>
              <td>{r.hours || '—'}</td>
              <td>{r.assign_role}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function WhatsApp({ lookups }) {
  const [rows, setRows] = useState([]);
  const [form, setForm] = useState({
    name: '',
    body: 'Hola {cliente}, te escribimos por tu servicio “{trabajo}” (orden #{orden}).',
    status_id: '',
    auto_open: false,
  });
  async function load() { setRows(await api('/api/whatsapp/templates')); }
  useEffect(() => { load().catch(console.error); }, []);
  const set = (k) => (e) => {
    const v = e.target.type === 'checkbox' ? e.target.checked : e.target.value;
    setForm((f) => ({ ...f, [k]: v }));
  };
  async function create(e) {
    e.preventDefault();
    await api('/api/whatsapp/templates', {
      method: 'POST',
      body: {
        ...form,
        status_id: form.status_id ? Number(form.status_id) : null,
      },
    });
    setForm({ name: '', body: form.body, status_id: '', auto_open: false });
    load();
  }
  async function toggleAuto(t) {
    await api(`/api/whatsapp/templates/${t.id}`, {
      method: 'PATCH',
      body: { auto_open: !t.auto_open },
    });
    load();
  }
  return (
    <section>
      <h2>WhatsApp con clientes</h2>
      <p className="meta">
        Hoy es viable con enlaces a WhatsApp (wa.me): abrís el chat con el texto ya armado y lo enviás vos.
        Un bot completo (respuesta automática sin abrir WhatsApp) es posible más adelante con la API oficial de Meta / WhatsApp Business (número verificado y plantillas aprobadas).
      </p>
      <p className="meta">
        Variables: {'{cliente}'} {'{trabajo}'} {'{estado}'} {'{fecha}'} {'{hora}'} {'{tecnico}'} {'{direccion}'} {'{orden}'}.
        Si marcás “abrir al cambiar estado”, al pasar una orden a ese estado te pregunta si abrir WhatsApp.
      </p>
      <form className="card grid" onSubmit={create} style={{ marginBottom: 16 }}>
        <div className="row">
          <Field label="Nombre plantilla"><input required value={form.name} onChange={set('name')} /></Field>
          <Field label="Estado que dispara (opcional)">
            <select value={form.status_id} onChange={set('status_id')}>
              <option value="">Mensaje libre</option>
              {lookups.statuses?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
          </Field>
          <label className="check">
            <input type="checkbox" checked={form.auto_open} onChange={set('auto_open')} />
            Abrir al cambiar a ese estado
          </label>
        </div>
        <Field label="Texto">
          <textarea required value={form.body} onChange={set('body')} rows={4} />
        </Field>
        <button className="primary" type="submit">Guardar plantilla</button>
      </form>
      <table>
        <thead>
          <tr><th>Plantilla</th><th>Estado</th><th>Auto</th><th>Texto</th><th></th></tr>
        </thead>
        <tbody>
          {rows.map((t) => (
            <tr key={t.id}>
              <td>{t.name}</td>
              <td>{t.status_name || 'Libre'}</td>
              <td>{t.auto_open ? 'Sí' : 'No'}</td>
              <td className="meta">{t.body}</td>
              <td>
                <button className="ghost" type="button" onClick={() => toggleAuto(t)}>
                  {t.auto_open ? 'Desactivar auto' : 'Activar auto'}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Usuarios() {
  const [rows, setRows] = useState([]);
  const [error, setError] = useState('');
  const [form, setForm] = useState({ username: '', password: '', full_name: '', role: 'coordinador' });
  async function load() {
    try {
      setError('');
      setRows(await api('/api/users'));
    } catch (err) {
      setError(err.message);
    }
  }
  useEffect(() => { load(); }, []);
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));
  async function create(e) {
    e.preventDefault();
    await api('/api/users', { method: 'POST', body: form });
    load();
  }
  return (
    <section>
      <h2>Usuarios y roles</h2>
      <p className="meta">Administrador: toda la app. Técnico: solo su agenda del día, editar/imprimir órdenes y hoja de ruta en Google Maps. Al crear un técnico se genera su perfil de agenda.</p>
      {error && <p className="error">{error}</p>}
      <form className="card row" onSubmit={create} style={{ marginBottom: 16 }}>
        <Field label="Usuario"><input required value={form.username} onChange={set('username')} /></Field>
        <Field label="Clave"><input required type="password" value={form.password} onChange={set('password')} /></Field>
        <Field label="Nombre"><input required value={form.full_name} onChange={set('full_name')} /></Field>
        <Field label="Rol">
          <select value={form.role} onChange={set('role')}>
            <option value="admin">Administrador (control total)</option>
            <option value="tecnico">Técnico (agenda propia)</option>
            <option value="coordinador">Coordinación</option>
          </select>
        </Field>
        <button className="primary" type="submit">Crear usuario</button>
      </form>
      <table>
        <thead><tr><th>Usuario</th><th>Nombre</th><th>Rol</th></tr></thead>
        <tbody>
          {rows.map((u) => (
            <tr key={u.id}><td>{u.username}</td><td>{u.full_name}</td><td>{ROLE_LABEL[u.role] || u.role}</td></tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

export default function App() {
  const [user, setUser] = useState(null);
  const [page, setPage] = useState('agenda');
  const [lookups, setLookups] = useState({});

  async function loadLookups() {
    setLookups(await api('/api/lookups'));
  }

  useEffect(() => {
    if (!getToken()) return;
    api('/api/auth/me')
      .then((d) => setUser(d.user))
      .catch(() => setToken(null));
  }, []);

  useEffect(() => {
    if (user) loadLookups().catch(console.error);
  }, [user]);

  useEffect(() => {
    if (!user) return;
    const allowed = PAGES.filter(([, , roles]) => roles.includes(user.role)).map(([id]) => id);
    if (!allowed.includes(page)) setPage('agenda');
  }, [page, user]);

  if (!user) return <Login onOk={setUser} />;

  const visible = PAGES.filter(([, , roles]) => roles.includes(user.role));

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">
          <img className="brand-logo" src="/logo-instal-service.jpg" alt="Instal Service S.A." />
          <small>Gestión de taller</small>
        </div>
        {visible.map(([id, label]) => (
          <button key={id} className={`nav-btn ${page === id ? 'active' : ''}`} onClick={() => setPage(id)}>
            {label}
          </button>
        ))}
        <div className="userbox">
          {user.full_name || user.username}<br />
          <span>{ROLE_LABEL[user.role] || user.role}</span>
          {user.technician_name && <><br /><span>{user.technician_name}</span></>}
          <button onClick={() => { setToken(null); setUser(null); }}>Salir</button>
        </div>
      </aside>
      <main className="main">
        {page === 'agenda' && <Agenda lookups={lookups} user={user} onLookups={loadLookups} />}
        {page === 'seguimientos' && <Seguimientos />}
        {page === 'ordenes' && <Ordenes lookups={lookups} onCreated={loadLookups} />}
        {page === 'clientes' && <Clientes lookups={lookups} onCreated={loadLookups} />}
        {page === 'tarifas' && <Tarifas lookups={lookups} />}
        {page === 'resultados' && <Resultados />}
        {page === 'catalogo' && <Catalogo lookups={lookups} onCreated={loadLookups} />}
        {page === 'tecnicos' && <Tecnicos onCreated={loadLookups} />}
        {page === 'proveedores' && <Proveedores onCreated={loadLookups} />}
        {page === 'reglas' && <Reglas lookups={lookups} />}
        {page === 'whatsapp' && <WhatsApp lookups={lookups} />}
        {page === 'usuarios' && <Usuarios />}
      </main>
    </div>
  );
}
