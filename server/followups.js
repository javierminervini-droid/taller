import { db } from './db.js';

function pickAssignee(role) {
  return db.prepare(
    'SELECT id FROM users WHERE role = ? AND active = 1 ORDER BY id LIMIT 1'
  ).get(role)?.id ?? null;
}

export function generateFollowups() {
  const rules = db.prepare('SELECT * FROM followup_rules WHERE active = 1').all();
  const insert = db.prepare(
    `INSERT INTO followups (service_order_id, rule_id, assigned_user_id, due_at, notes, status)
     VALUES (?, ?, ?, datetime('now'), ?, 'pendiente')`
  );
  const exists = db.prepare(
    `SELECT id FROM followups WHERE service_order_id = ? AND rule_id = ? AND status = 'pendiente'`
  );

  for (const rule of rules) {
    let orders = [];
    if (rule.trigger_type === 'status') {
      orders = db.prepare(
        `SELECT o.* FROM service_orders o
         JOIN service_statuses s ON s.id = o.status_id
         WHERE o.status_id = ? AND s.is_closed = 0`
      ).all(rule.status_id);
    } else if (rule.trigger_type === 'elapsed_hours' && rule.status_id && rule.hours) {
      orders = db.prepare(
        `SELECT o.* FROM service_orders o
         JOIN service_statuses s ON s.id = o.status_id
         WHERE o.status_id = ? AND s.is_closed = 0
           AND datetime(COALESCE(o.started_at, o.created_at)) <= datetime('now', ?) `
      ).all(rule.status_id, `-${rule.hours} hours`);
    }

    for (const order of orders) {
      if (exists.get(order.id, rule.id)) continue;
      insert.run(order.id, rule.id, pickAssignee(rule.assign_role), rule.name);
    }
  }
}

export function listFollowups(status = 'pendiente') {
  generateFollowups();
  return db.prepare(
    `SELECT f.*, o.title AS order_title, o.scheduled_date, c.name AS client_name,
            u.full_name AS assigned_name, r.name AS rule_name, s.name AS order_status
     FROM followups f
     JOIN service_orders o ON o.id = f.service_order_id
     JOIN clients c ON c.id = o.client_id
     JOIN service_statuses s ON s.id = o.status_id
     LEFT JOIN users u ON u.id = f.assigned_user_id
     LEFT JOIN followup_rules r ON r.id = f.rule_id
     WHERE (? = 'todos' OR f.status = ?)
     ORDER BY f.status ASC, f.created_at DESC`
  ).all(status, status);
}
