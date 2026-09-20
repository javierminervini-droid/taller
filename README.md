# Taller Gestión

Sistema liviano para servicio técnico y venta de repuestos. Corre en Windows 10 o posterior con Node.js, usa SQLite (un archivo local) y se abre en el navegador. Mañana la misma API puede usarse desde Android e iOS.

## Cómo arrancar

1. Doble clic en `iniciar.bat`, o en la carpeta del proyecto:

```bat
npm install
npm run build --prefix client
npm start
```

2. Abrí **solo** esta dirección: http://127.0.0.1:3847

No uses `5173` ni `5180` (quedan en blanco o desactualizados).

Usuarios de prueba:

- `admin` / `admin123`
- `coord` / `coord123`
- `diego` / `diego123`

## Qué incluye

- Clientes con origen en **proveedores** o **prestadores** (bases distintas, con ID externo).
- Agenda diaria por técnico, con filtros por tipo de producto, proveedor, técnico, fecha y localidad.
- Órdenes de servicio con estados.
- Seguimientos automáticos al entrar a un estado o al cumplir X horas en ese estado.
- Usuarios con roles: admin, coordinador, técnico.

## Producción local

```bat
npm run build
npm start
```

Queda en `http://localhost:3847` y escucha en la red local (`0.0.0.0`) para entrar desde un celular en la misma Wi‑Fi.
