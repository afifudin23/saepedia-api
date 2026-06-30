# Cross-platform Makefile.
# Windows  → recipe berbasis PowerShell.
# Linux/macOS → recipe berbasis sh (default).
ifeq ($(OS),Windows_NT)
SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command
endif

ifneq (,$(wildcard .env))
include .env
export
endif

BIN_NAME := saepedia-api
DB_HOST  ?= localhost
DB_PORT  ?= 5432
DB_USER  ?= postgres
DB_PASSWORD ?=
DB_NAME  ?= seapedia_db

.PHONY: setup tidy run air build seed seed-one release swag migrate-up migrate-down migrate-drop migrate-reset db-reset migrate-create migrate-version migrate-force

# ── Perintah lintas-OS (pakai tooling Go, sama di mana saja) ────────────────
setup:
	go mod tidy
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest

tidy:
	go mod tidy

run:
	go run ./cmd/api

air:
	air -c .air.toml

# Generate Swagger docs (docs/swagger) dari anotasi handler
swag:
	swag init -g cmd/api/main.go -o docs/swagger --parseInternal

seed:
	go run ./scripts/seed

release:
	go run ./scripts/release

# ════════════════════════════════════════════════════════════════════════════
ifeq ($(OS),Windows_NT)
# ── Windows (PowerShell) ────────────────────────────────────────────────────
build:
	go build -o bin/$(BIN_NAME).exe ./cmd/api

seed-one:
	@if (-not '$(name)') { throw 'name is required. Example: make seed-one name=user' }; go run ./scripts/seed data=$(name)

migrate-up:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url up

migrate-down:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url down $(if $(n),$(n),1)

migrate-drop:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url drop -f

migrate-reset:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url drop -f; migrate -path migrations -database $$url up

db-reset:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url drop -f; migrate -path migrations -database $$url up; go run ./scripts/seed

migrate-create:
	@if (-not '$(name)') { throw 'name is required. Example: make migrate-create name=create_users_table' }; migrate create -ext sql -dir migrations -format '20060102150405' $(name)

migrate-version:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url version

migrate-force:
	@if (-not '$(version)') { throw 'version is required. Example: make migrate-force version=1' }; $$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url force $(version)

else
# ── Linux / macOS (sh) ──────────────────────────────────────────────────────
# Catatan: password disisipkan apa adanya. Bila mengandung karakter spesial
# (mis. / @ : ?), URL-encode dulu di .env (mis. "/" → "%2F").
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

build:
	go build -o bin/$(BIN_NAME) ./cmd/api

seed-one:
	@if [ -z "$(name)" ]; then echo "name is required. Example: make seed-one name=user"; exit 1; fi; go run ./scripts/seed data=$(name)

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down $(if $(n),$(n),1)

migrate-drop:
	migrate -path migrations -database "$(DB_URL)" drop -f

migrate-reset:
	migrate -path migrations -database "$(DB_URL)" drop -f && migrate -path migrations -database "$(DB_URL)" up

db-reset:
	migrate -path migrations -database "$(DB_URL)" drop -f && migrate -path migrations -database "$(DB_URL)" up && go run ./scripts/seed

migrate-create:
	@if [ -z "$(name)" ]; then echo "name is required. Example: make migrate-create name=create_users_table"; exit 1; fi; migrate create -ext sql -dir migrations -format '20060102150405' $(name)

migrate-version:
	migrate -path migrations -database "$(DB_URL)" version

migrate-force:
	@if [ -z "$(version)" ]; then echo "version is required. Example: make migrate-force version=1"; exit 1; fi; migrate -path migrations -database "$(DB_URL)" force $(version)

endif
