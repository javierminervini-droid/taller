import * as XLSX from 'xlsx';
import { db } from './db.js';

export const TEMPLATE_HEADERS = [
  'ID_Salesforce',
  'Nombre',
  'Telefono',
  'Email',
  'Direccion',
  'Localidad',
  'Provincia',
  'Notas',
];

const ALIASES = {
  ID_Salesforce: ['id_salesforce', 'id', 'salesforce id', 'account id', 'accountid', 'case number', 'casenumber', 'id externo', 'external id'],
  Nombre: ['nombre', 'name', 'account name', 'accountname', 'razon social', 'cliente'],
  Telefono: ['telefono', 'phone', 'mobile', 'mobilephone', 'celular'],
  Email: ['email', 'correo'],
  Direccion: ['direccion', 'dirección', 'street', 'billingstreet', 'address', 'domicilio'],
  Localidad: ['localidad', 'city', 'billingcity', 'ciudad'],
  Provincia: ['provincia', 'state', 'billingstate'],
  Notas: ['notas', 'notes', 'description', 'descripcion', 'descripción'],
};

function norm(value) {
  return String(value || '')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .trim()
    .toLowerCase();
}

function mapRow(row) {
  const keys = Object.keys(row);
  const pick = (field) => {
    const aliases = ALIASES[field];
    const found = keys.find((k) => aliases.includes(norm(k)));
    const val = found ? row[found] : '';
    return val == null ? '' : String(val).trim();
  };
  return {
    external_id: pick('ID_Salesforce'),
    name: pick('Nombre'),
    phone: pick('Telefono'),
    email: pick('Email'),
    address: pick('Direccion'),
    locality: pick('Localidad'),
    province: pick('Provincia'),
    notes: pick('Notas'),
  };
}

function localityText(name, province) {
  if (!name) return null;
  if (province) return `${name}, ${province}`;
  return name;
}

export function buildTemplateBuffer() {
  const sample = [{
    ID_Salesforce: '001XX000003ABC',
    Nombre: 'Juan Pérez',
    Telefono: '11 5555-0000',
    Email: 'juan@correo.com',
    Direccion: 'Av. Siempre Viva 742',
    Localidad: 'Quilmes',
    Provincia: 'Buenos Aires',
    Notas: 'Exportado desde Salesforce',
  }];
  const sheet = XLSX.utils.json_to_sheet(sample, { header: TEMPLATE_HEADERS });
  const wb = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(wb, sheet, 'Clientes');
  return XLSX.write(wb, { type: 'buffer', bookType: 'xlsx' });
}

export function importClientsFromBuffer(buffer, providerId) {
  const wb = XLSX.read(buffer, { type: 'buffer' });
  const sheet = wb.Sheets[wb.SheetNames[0]];
  const rows = XLSX.utils.sheet_to_json(sheet, { defval: '' });
  const find = db.prepare(
    'SELECT id FROM clients WHERE provider_id = ? AND external_id = ? LIMIT 1'
  );
  const insert = db.prepare(
    `INSERT INTO clients (provider_id, external_id, name, phone, email, locality, address, notes, source)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'excel')`
  );
  const update = db.prepare(
    `UPDATE clients SET name=?, phone=?, email=?, locality=?, address=?, notes=?, source='excel'
     WHERE id=?`
  );

  let created = 0;
  let updated = 0;
  const errors = [];

  for (const [i, raw] of rows.entries()) {
    const row = mapRow(raw);
    if (!row.name) {
      errors.push(`Fila ${i + 2}: falta Nombre`);
      continue;
    }
    const locality = localityText(row.locality, row.province);
    const ext = row.external_id || null;
    const existing = ext ? find.get(providerId, ext) : null;
    if (existing) {
      update.run(row.name, row.phone || null, row.email || null, locality, row.address || null, row.notes || null, existing.id);
      updated += 1;
    } else {
      insert.run(providerId, ext, row.name, row.phone || null, row.email || null, locality, row.address || null, row.notes || null);
      created += 1;
    }
  }

  return { created, updated, errors, total: rows.length };
}
