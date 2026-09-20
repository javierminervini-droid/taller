# Legacy Node API (reference only)

This Fastify + SQLite server was the original monolith. The default stack is now:

- Go API: `backend/` (Gin + Bun + PostgreSQL)
- React client: `client/`

All former Phase-2 features (Excel import, WhatsApp, tariffs/results) live in the Go API.

To run this legacy server (SQLite file in `data/`) only for historical comparison:

```bash
npm run legacy:server
```

Do not use it for new features.
