# Taller Gestión

Sistema liviano para servicio técnico y venta de repuestos. Stack actual:

- **Frontend:** React + Vite (`client/`)
- **Backend:** Go API con Gin + Bun (`backend/`, código en `backend/src/`)
- **DB:** PostgreSQL 16 vía Docker Compose

## Requisitos

- [Go](https://go.dev/dl/) 1.22+
- [Docker](https://docs.docker.com/get-docker/) (Compose v2)
- Node.js 20+ (solo para el cliente React)
- [Make](https://gnuwin32.sourceforge.net/packages/make.htm) o Git Bash / WSL (recomendado en Windows)

## Arranque rápido (otros devs)

```bash
cp .env.example .env
make up          # Postgres + API en Docker (puerto 3847)
# o desarrollo local del API:
make dev         # levanta solo Postgres
make api         # terminal 1 — Go API en :3847
make frontend    # terminal 2 — Vite en :5180
```

Abrí **http://127.0.0.1:5180** en desarrollo (Vite proxy → API), o **http://127.0.0.1:3847** si usás solo el contenedor `api`.

Usuarios de prueba:

- `admin` / `admin123`
- `coord` / `coord123`
- `diego` / `diego123`

Más detalle: [docs/dev-setup.md](docs/dev-setup.md).

## Makefile

| Comando | Qué hace |
| --- | --- |
| `make up` | Compose: DB + API |
| `make down` | Para contenedores |
| `make db-reset` | Borra volumen Postgres, migra y seed |
| `make migrate` / `make seed` | Schema / datos demo |
| `make api` / `make frontend` | Procesos en el host |
| `make qa` | Smoke tests contra la API |

## Qué incluye

- Clientes, proveedores/prestadores, agenda, órdenes de servicio, seguimientos, usuarios/roles
- Auth JWT (`admin`, `coordinador`, `tecnico`)
- Import Excel de clientes / plantilla Salesforce
- WhatsApp templates y mensajes `wa.me`
- Tarifas y reporte de resultados (ingresos/costos)

El servidor Node legacy queda en `server/` (`npm run legacy:server`) solo como referencia histórica.

## Legacy (SQLite + Node)

```bat
npm run legacy:server
```

Usa el archivo SQLite en `data/` y no requiere Docker. No es el camino por defecto.
