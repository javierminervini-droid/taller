# Variables (override via environment)
COMPOSE ?= docker compose
BACKEND_DIR := backend
CLIENT_DIR := client
ENV_FILE := .env

.PHONY: help env up down logs db-reset migrate seed reseed api frontend dev build test tidy qa

help:
	@echo "Targets:"
	@echo "  make env       - copy .env.example to .env if missing"
	@echo "  make up        - start Postgres + API (Docker)"
	@echo "  make down      - stop containers"
	@echo "  make logs      - follow compose logs"
	@echo "  make migrate   - apply SQL migrations"
	@echo "  make seed      - seed demo data if empty"
	@echo "  make reseed    - wipe app tables and reload demo data"
	@echo "  make db-reset  - wipe Postgres volume, migrate, seed"
	@echo "  make api       - run Go API on host against compose DB"
	@echo "  make frontend  - run Vite React client"
	@echo "  make dev       - start DB container, then print how to run api+frontend"
	@echo "  make build     - build Go API binary + client dist"
	@echo "  make test      - run Go tests"
	@echo "  make tidy      - go mod tidy"
	@echo "  make qa        - smoke-test API (API must be running)"

env:
	@test -f $(ENV_FILE) || cp .env.example $(ENV_FILE)

up: env
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

db-reset: env
	$(COMPOSE) down -v
	$(COMPOSE) up -d db
	@echo "Waiting for Postgres..."
	@$(COMPOSE) exec -T db sh -c 'until pg_isready -U "$${POSTGRES_USER:-taller}"; do sleep 1; done'
	cd $(BACKEND_DIR) && go run ./cmd/api -seed-only
	@echo "Database reset complete."

migrate:
	cd $(BACKEND_DIR) && go run ./cmd/api -migrate-only

seed:
	cd $(BACKEND_DIR) && go run ./cmd/api -seed-only

reseed:
	cd $(BACKEND_DIR) && go run ./cmd/api -force-seed

api: env
	cd $(BACKEND_DIR) && go run ./cmd/api

frontend:
	npm run dev --prefix $(CLIENT_DIR)

dev: env
	$(COMPOSE) up -d db
	@echo ""
	@echo "Postgres is up. In two terminals run:"
	@echo "  make api"
	@echo "  make frontend"
	@echo "Then open http://127.0.0.1:5180"
	@echo ""

build:
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api
	npm run build --prefix $(CLIENT_DIR)

test:
	cd $(BACKEND_DIR) && go test ./...

tidy:
	cd $(BACKEND_DIR) && go mod tidy

qa:
	node scripts/qa.mjs
