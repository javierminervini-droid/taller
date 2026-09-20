import { db } from './db.js';

function money(n) {
  return Math.round(Number(n || 0) * 100) / 100;
}

export function listTariffs() {
  return db.prepare(
    `SELECT t.*, p.name AS provider_name, tech.name AS technician_name
     FROM tariffs t
     LEFT JOIN providers p ON p.id = t.provider_id
     LEFT JOIN technicians tech ON tech.id = t.technician_id
     ORDER BY t.scope, t.id`
  ).all();
}

function pickIncome(order, tariffs) {
  return tariffs.find((t) => t.active && t.scope === 'proveedor' && t.provider_id === order.provider_id)
    || tariffs.find((t) => t.active && t.scope === 'general')
    || {};
}

function pickCost(order, tariffs) {
  return tariffs.find((t) => t.active && t.scope === 'tecnico' && t.technician_id === order.technician_id)
    || tariffs.find((t) => t.active && t.scope === 'tecnico' && !t.technician_id)
    || tariffs.find((t) => t.active && t.scope === 'general')
    || {};
}

export function quoteOrder(order, tariffs) {
  const hours = Number(order.hours || 0);
  const km = Number(order.km || 0);
  const partsCost = Number(order.parts_cost || 0);
  const partsSale = Number(order.parts_sale || 0);
  const incomeT = pickIncome(order, tariffs);
  const costT = pickCost(order, tariffs);
  const income = money(
    Number(incomeT.income_fixed || 0)
    + hours * Number(incomeT.income_per_hour || 0)
    + km * Number(incomeT.income_per_km || 0)
    + partsSale * Number(incomeT.income_parts_pct || 0) / 100
  );
  const cost = money(
    Number(costT.cost_fixed || 0)
    + hours * Number(costT.cost_per_hour || 0)
    + km * Number(costT.cost_per_km || 0)
    + partsCost * Number(costT.cost_parts_pct || 0) / 100
  );
  return {
    income,
    cost,
    profit: money(income - cost),
    income_tariff: incomeT.name || null,
    cost_tariff: costT.name || null,
  };
}

export function resultsReport({ from, to, group = 'proveedor' }) {
  const tariffs = listTariffs();
  const orders = db.prepare(
    `SELECT o.*, c.name AS client_name, t.name AS technician_name,
            pr.name AS provider_name, pr.kind AS provider_kind
     FROM service_orders o
     JOIN clients c ON c.id = o.client_id
     LEFT JOIN technicians t ON t.id = o.technician_id
     LEFT JOIN providers pr ON pr.id = o.provider_id
     WHERE (? IS NULL OR o.scheduled_date >= ?)
       AND (? IS NULL OR o.scheduled_date <= ?)
     ORDER BY o.scheduled_date, o.id`
  ).all(from || null, from || null, to || null, to || null);

  const quoted = orders.map((o) => ({ ...o, ...quoteOrder(o, tariffs) }));
  const buckets = new Map();
  for (const o of quoted) {
    let key = 'sin-asignar';
    let label = 'Sin asignar';
    if (group === 'cliente') {
      key = String(o.client_id);
      label = o.client_name;
    } else if (group === 'tecnico') {
      key = String(o.technician_id || 0);
      label = o.technician_name || 'Sin técnico';
    } else {
      key = String(o.provider_id || 0);
      label = o.provider_name ? `${o.provider_name} (${o.provider_kind})` : 'Carga propia / sin origen';
    }
    if (!buckets.has(key)) {
      buckets.set(key, { key, label, orders: 0, income: 0, cost: 0, profit: 0 });
    }
    const b = buckets.get(key);
    b.orders += 1;
    b.income = money(b.income + o.income);
    b.cost = money(b.cost + o.cost);
    b.profit = money(b.profit + o.profit);
  }

  const rows = [...buckets.values()].sort((a, b) => b.profit - a.profit);
  const totals = rows.reduce(
    (acc, r) => ({
      orders: acc.orders + r.orders,
      income: money(acc.income + r.income),
      cost: money(acc.cost + r.cost),
      profit: money(acc.profit + r.profit),
    }),
    { orders: 0, income: 0, cost: 0, profit: 0 }
  );
  return { group, from, to, rows, totals, orders: quoted };
}
