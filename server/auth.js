import jwt from 'jsonwebtoken';
import bcrypt from 'bcryptjs';
import { db } from './db.js';

const SECRET = process.env.JWT_SECRET || 'taller-gestion-local-dev';

export function signUser(user) {
  return jwt.sign(
    { id: user.id, username: user.username, role: user.role, full_name: user.full_name },
    SECRET,
    { expiresIn: '12h' }
  );
}

export function verifyToken(token) {
  return jwt.verify(token, SECRET);
}

export function login(username, password) {
  const user = db.prepare('SELECT * FROM users WHERE username = ? AND active = 1').get(username);
  if (!user || !bcrypt.compareSync(password, user.password_hash)) return null;
  return user;
}

export function requireAuth(request, reply) {
  const header = request.headers.authorization || '';
  const token = header.startsWith('Bearer ') ? header.slice(7) : null;
  if (!token) {
    reply.code(401).send({ error: 'Sesión requerida' });
    return null;
  }
  try {
    request.user = verifyToken(token);
    return request.user;
  } catch {
    reply.code(401).send({ error: 'Sesión inválida' });
    return null;
  }
}

export function requireRole(user, roles, reply) {
  if (!roles.includes(user.role)) {
    reply.code(403).send({ error: 'Sin permiso para esta acción' });
    return false;
  }
  return true;
}

export function isAdmin(user) {
  return user.role === 'admin';
}

export function canManage(user) {
  return user.role === 'admin' || user.role === 'coordinador';
}

export function technicianOf(user) {
  if (!user || user.role !== 'tecnico') return null;
  return db.prepare('SELECT * FROM technicians WHERE user_id = ? AND active = 1').get(user.id) || null;
}

export function publicUser(user) {
  const tech = technicianOf(user) || (user.role === 'tecnico'
    ? db.prepare('SELECT * FROM technicians WHERE user_id = ?').get(user.id)
    : null);
  return {
    id: user.id,
    username: user.username,
    full_name: user.full_name,
    role: user.role,
    technician_id: tech?.id ?? null,
    technician_name: tech?.name ?? null,
  };
}
