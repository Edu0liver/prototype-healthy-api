.PHONY: run build test lint ci migrate-up migrate-down migrate-create swag swag-check seed deps docker-build

-include .env
export

BINARY=prototype-healthy-api
MAIN=./cmd/api
MIGRATIONS_DIR=./migrations
GOOSE=$(shell go env GOPATH)/bin/goose
SWAG=$(shell go env GOPATH)/bin/swag
GOOSE_DRIVER=postgres
# Migrations run as the superuser; fall back to DATABASE_URL if unset.
GOOSE_DBSTRING?=$(or $(MIGRATE_DATABASE_URL),$(DATABASE_URL))

VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# ── Dev ────────────────────────────────────────────────────────────────────────

run:
	go run $(MAIN)/main.go

build:
	go build \
	  -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)" \
	  -o bin/$(BINARY) $(MAIN)

deps:
	go mod download
	go mod tidy

# ── Lint ───────────────────────────────────────────────────────────────────────

lint:
	go vet ./...
	@gofmt -l . | grep -v '^docs/' | (! grep .) || (echo "gofmt: files need formatting (run: gofmt -w .)" && exit 1)

# ── Tests ──────────────────────────────────────────────────────────────────────

test:
	go test ./... -v -race -count=1

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ── Migrations ────────────────────────────────────────────────────────────────

migrate-up:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" $(GOOSE) -dir $(MIGRATIONS_DIR) up

migrate-down:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" $(GOOSE) -dir $(MIGRATIONS_DIR) down

migrate-status:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" $(GOOSE) -dir $(MIGRATIONS_DIR) status

migrate-create:
	@read -p "Migration name: " name; \
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" $(GOOSE) -dir $(MIGRATIONS_DIR) create $${name} sql

# ── Swagger ────────────────────────────────────────────────────────────────────

swag:
	$(SWAG) init -g cmd/api/main.go -o docs --parseDependency --parseInternal

swag-check:
	@tmp=$$(mktemp -d); \
	$(SWAG) init -g cmd/api/main.go -o $$tmp --parseDependency --parseInternal; \
	diff -q $$tmp/swagger.json docs/swagger.json && diff -q $$tmp/swagger.yaml docs/swagger.yaml || \
	  (echo "swagger docs out of date — run make swag" && rm -rf $$tmp && exit 1); \
	rm -rf $$tmp

# ── CI (mirrors GitHub Actions) ────────────────────────────────────────────────

ci: lint swag-check build test

# ── Docker ─────────────────────────────────────────────────────────────────────

docker-build:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILD_DATE=$(BUILD_DATE) \
	  -t lumia-api:$(VERSION) .

up:
	docker compose up -d

up-db:
	docker compose up -d postgres redis evolution minio

down:
	docker compose down

# ── Seed ───────────────────────────────────────────────────────────────────────

seed:
	go run ./scripts/seed/main.go
