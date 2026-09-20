# Dev setup — Taller Gestión

## One-time setup

1. Install **Go 1.22+**, **Docker Desktop** (or Engine + Compose), **Node.js 20+**, and **Make**.
2. Clone the repo and copy env:

```bash
cp .env.example .env
```

3. Pull deps:

```bash
cd backend && go mod tidy && cd ..
npm install
```

## Daily workflow

**Option A — full Docker (API + DB):**

```bash
make up
# API: http://127.0.0.1:3847/health
make frontend   # UI on :5180 with /api proxy
```

**Option B — DB in Docker, API on host (faster iteration):**

```bash
make dev        # starts Postgres only
make api        # terminal 1
make frontend   # terminal 2
```

## Schema bootstrap

Migrations live in [`backend/migrations/`](../backend/migrations/). Application code is under [`backend/src/`](../backend/src/) (`controllers` → `services` → `repositories`). On API start migrations are applied automatically. Demo users/data seed when `RUN_SEED=true` and the `users` table is empty.

```bash
make db-reset   # wipe volume + migrate + seed
make migrate    # schema only
make seed       # demo data if empty
```

## Smoke test

With the API running:

```bash
make qa
```

QA covers auth, orders, followups, WhatsApp templates, tariffs, and the Excel client template.

## Demo users

| User | Password | Role |
| --- | --- | --- |
| admin | admin123 | admin |
| coord | coord123 | coordinador |
| diego | diego123 | tecnico |
| sofia | sofia123 | tecnico |
