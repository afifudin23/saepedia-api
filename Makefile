SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command

ifneq (,$(wildcard .env))
include .env
export
endif

APP_NAME ?= saepedia-api
DB_HOST  ?= localhost
DB_PORT  ?= 5432
DB_USER  ?= postgres
DB_PASSWORD ?=
DB_NAME  ?= seapedia_db

.PHONY: setup tidy run air build seed seed-one release swag migrate-up migrate-down migrate-drop migrate-reset db-reset migrate-create migrate-version migrate-force

setup:
	go mod tidy
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs (docs/swagger) dari anotasi handler
swag:
	swag init -g cmd/api/main.go -o docs/swagger --parseInternal

tidy:
	go mod tidy

run:
	go run ./cmd/api

air:
	air -c .air.toml

build:
	go build -o bin/$(APP_NAME).exe ./cmd/api

seed:
	go run ./scripts/seed

seed-one:
	@if (-not '$(name)') { throw 'name is required. Example: make seed-one name=user' }; go run ./scripts/seed data=$(name)

release:
	go run ./scripts/release

migrate-up:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url up

# Mundur N langkah (default 1). Contoh: make migrate-down n=3
migrate-down:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url down $(if $(n),$(n),1)

# Hapus SEMUA tabel (kembali ke nol, tanpa prompt)
migrate-drop:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url drop -f

# Reset skema dari nol: drop semua lalu migrate up lagi
migrate-reset:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url drop -f; migrate -path migrations -database $$url up

# Reset penuh + isi ulang data demo (drop -> up -> seed)
db-reset:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url drop -f; migrate -path migrations -database $$url up; go run ./scripts/seed

migrate-create:
	@if (-not '$(name)') { throw 'name is required. Example: make migrate-create name=create_users_table' }; migrate create -ext sql -dir migrations -format '20060102150405' $(name)

migrate-version:
	@$$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url version

migrate-force:
	@if (-not '$(version)') { throw 'version is required. Example: make migrate-force version=1' }; $$pwd = [uri]::EscapeDataString('$(DB_PASSWORD)'); $$url = "postgres://$(DB_USER):$$pwd@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"; migrate -path migrations -database $$url force $(version)
