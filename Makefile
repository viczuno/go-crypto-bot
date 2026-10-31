SHELL := /bin/sh
.PHONY: install dev db-up db-migrate import-history verify-history ingest test test-go test-web build lint fmt
install:
	go mod download
	npm --prefix web ci
dev:
	npm --prefix web run dev
db-up:
	docker compose up -d --wait
db-migrate:
	go run ./cmd/db
import-history:
	go run ./cmd/import-history -mode import
verify-history:
	go run ./cmd/import-history -mode verify
ingest:
	go run ./cmd/ingest
test: test-go test-web
test-go:
	go test -race ./cmd/... ./internal/...
test-web:
	npm --prefix web test
build:
	go build ./cmd/... ./internal/...
	npm --prefix web run build
lint:
	go vet ./cmd/... ./internal/...
	npm --prefix web run lint
	npm --prefix web run typecheck
	npm --prefix web run format:check
fmt:
	gofmt -w cmd internal
	npm --prefix web run format
