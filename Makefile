# Muat variabel dari .env jika file-nya ada.
-include .env
export

GOOSE := go tool goose -dir migrations postgres "$(DATABASE_URL)"

.DEFAULT_GOAL := help
.PHONY: help setup db-up db-down migrate-up migrate-down migrate-status migrate-new run test cover lint fmt

help: ## Tampilkan daftar perintah
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

setup: ## Download dependency dan pasang goose sebagai Go tool
	go get github.com/jackc/pgx/v5
	go get -tool github.com/pressly/goose/v3/cmd/goose
	go mod tidy

db-up: ## Jalankan PostgreSQL di Docker
	docker compose up -d --wait

db-down: ## Matikan PostgreSQL (data tetap tersimpan di volume)
	docker compose down

migrate-up: ## Jalankan semua migration yang belum diterapkan
	$(GOOSE) up

migrate-down: ## Batalkan migration terakhir
	$(GOOSE) down

migrate-status: ## Lihat status migration
	$(GOOSE) status

migrate-new: ## Buat file migration baru, contoh: make migrate-new name=create_tenants
	go tool goose -dir migrations -s create $(name) sql

run: ## Jalankan API
	go run ./cmd/api

test: ## Jalankan semua test
	go test ./...

cover: ## Jalankan test dengan laporan coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Jalankan linter
	golangci-lint run ./...

fmt: ## Rapikan format kode
	golangci-lint fmt ./...
