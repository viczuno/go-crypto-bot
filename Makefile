SHELL := /bin/sh
.DEFAULT_GOAL := help
.PHONY: help install dev db-up db-migrate import-history verify-history ingest test test-go test-web build lint fmt
install: ## Install locked dependencies
	go mod download
	npm --prefix web ci
dev: ## Start Next.js
	npm --prefix web run dev
db-up: ## Start optional Docker PostgreSQL
	docker compose up -d --wait
db-migrate: ## Apply schema to configured direct DB
	go run ./cmd/db
import-history: ## Import immutable recovery archive
	go run ./cmd/import-history -mode import
verify-history: ## Verify archive against configured DB
	go run ./cmd/import-history -mode verify
ingest: ## Run configured Go ingestion
	go run ./cmd/ingest
test: test-go test-web ## Run Go and frontend tests
test-go: ## Run Go race tests; DB tests opt in
	go test -race ./cmd/... ./internal/...
test-web: ## Run frontend tests
	npm --prefix web test
build: ## Build Go and Next.js
	go build ./cmd/... ./internal/...
	npm --prefix web run build
lint: ## Check Go, frontend types, lint and format
	go vet ./cmd/... ./internal/...
	npm --prefix web run lint
	npm --prefix web run typecheck
	npm --prefix web run format:check
fmt: ## Format Go and frontend source
	gofmt -w cmd internal
	npm --prefix web run format

help: ## List developer commands
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "%-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
