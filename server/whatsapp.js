import { db } from './db.js';

/** Normaliza teléfonos AR a formato internacional sin + (54911...) */
export function normalizePhone(raw) {
  if (!raw) return null;
  let digits = String(raw).replace(/\D/g, '');
  if (!digits) return null;
  if (digits.startsWith('00')) digits = digits.slice(2);
  if (digits.startsWith('54')) {
    // ok
  } else if (digits.startsWith('0')) {
    digits = `54${digits.slice(1)}`;
  } else if (digits.length <= 10) {
    digits = `54${digits}`;
  }
  // WhatsApp AR móvil suele ir con 9 después del 54
  if (digits.startsWith('54') && !digits.startsWith('549') && digits.length >= 12) {
    digits = `549${digits.slice(2)}`;
  }
  return digits;
}

export function telHref(raw) {
  const digits = String(raw || '').replace(/\D/g, '');
  if (!digits) return null;
  return `tel:${digits}`;
}

function fillTemplate(body, vars) {
  return String(body || '').replace(/\{(\w+)\}/g, (_, key) => {
    const v = vars[key];
    return v == null || v === '' ? '' : String(v);
  });
}

export function orderVars(order) {
  return {
    cliente: order.client_name || '',
    trabajo: order.title || '',
    estado: order.status_name || '',
    fecha: order.scheduled_date || '',
    hora: order.scheduled_time || '',
    tecnico: order.technician_name || '',
    direccion: [order.client_address, order.locality].filter(Boolean).join(', '),
    telefono: order.client_phone || '',
    orden: String(order.id || ''),
  };
}

export function buildWhatsAppUrl(phone, text) {
  const n = normalizePhone(phone);
  if (!n) return null;
  return `https://wa.me/${n}?text=${encodeURIComponent(text || '')}`;
}

export function listWhatsAppTemplates() {
  return db.prepare(
    `SELECT t.*, s.name AS status_name
     FROM whatsapp_templates t
     LEFT JOIN service_statuses s ON s.id = t.status_id
     ORDER BY t.id`
  ).all();
}

export function messageForOrder(order, templateId = null) {
  let tpl = null;
  if (templateId) {
    tpl = db.prepare('SELECT * FROM whatsapp_templates WHERE id = ? AND active = 1').get(templateId);
  } else if (order.status_id) {
    tpl = db.prepare(
      `SELECT * FROM whatsapp_templates
       WHERE active = 1 AND status_id = ?
       ORDER BY id LIMIT 1`
    ).get(order.status_id);
  }
  if (!tpl) {
    tpl = db.prepare(
      `SELECT * FROM whatsapp_templates WHERE active = 1 AND status_id IS NULL ORDER BY id LIMIT 1`
    ).get();
  }
  if (!tpl) return null;
  const text = fillTemplate(tpl.body, orderVars(order));
  const url = buildWhatsAppUrl(order.client_phone, text);
  if (!url) return { error: 'El cliente no tiene teléfono válido', template: tpl.name, text };
  return { url, text, template: tpl.name, template_id: tpl.id, auto_open: !!tpl.auto_open };
}
