# xermess — run `make` for the list of targets.

SHELL   := bash
APPS    := console id
# Not ./...: that walks into web/*/node_modules.
PKGS    := ./cmd/... ./internal/... ./migrations/... ./i18n/...
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMPOSE := docker compose -f deploy/compose.yaml
# The same stack on one machine: .localhost hostnames, Caddy's own CA. An
# explicit -f means compose picks up no override on its own, so it is named.
COMPOSE_LOCAL := $(COMPOSE) -f deploy/compose.local.yaml

# `make full-start prod` and `make full-start -- --prod` both select prod; make
# itself rejects a bare `--prod` as one of its own options.
MODE ?= $(if $(filter prod --prod,$(MAKECMDGOALS)),prod,dev)

.DEFAULT_GOAL := help
.PHONY: help setup full-start dev prod --dev --prod run build test test-integration check \
	migrate-up migrate-down migrate-status migrate-new db-create db-reset db-psql \
	web-check web-build deploy-build deploy-up deploy-down deploy-logs \
	deploy-local deploy-local-down deploy-local-logs clean

help: ## Show this help
	@awk -F ':.*## ' '/^## / {printf "\n\033[1m%s\033[0m\n", substr($$0, 4)} /^[a-z -]+:.*## / {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Setup

setup: ## Prepare a fresh checkout: .env, database, dependencies
	@scripts/setup.sh

## Run

full-start: ## Start the API and all web apps — make full-start dev|prod
	@scripts/start.sh --$(MODE)

ifneq (,$(filter full-start,$(MAKECMDGOALS)))
dev prod --dev --prod: ; @:
else
dev: ## Same as make full-start dev
	@scripts/start.sh --dev
prod: ## Same as make full-start prod
	@scripts/start.sh --prod
endif

run: ## Start only the API
	go run ./cmd/xermess

build: ## Build the API into bin/xermess
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/xermess ./cmd/xermess

## Quality

check: ## Format, vet and test the Go code
	go fmt $(PKGS) && go vet $(PKGS) && go test $(PKGS)

test: ## Run the Go tests
	go test $(PKGS)

test-integration: ## Run the tests that need Postgres and Redis (throwaway databases and key prefixes)
	@. scripts/lib.sh && load_env && XERMESS_TEST_DB_DSN="$${DB_URL:-$$XERMESS_DB_DSN}" \
		XERMESS_TEST_REDIS="$${XERMESS_TEST_REDIS:-$${XERMESS_REDIS_HOST:+$$XERMESS_REDIS_HOST:$${XERMESS_REDIS_PORT:-6379}}}" \
		go test -count=1 -run Live ./internal/...

web-check: ## Lint and type-check the web apps
	@for app in $(APPS); do echo "==> $$app"; (cd web/$$app && bun run lint && bun run check) || exit 1; done

web-build: ## Build the web apps
	@for app in $(APPS); do echo "==> $$app"; (cd web/$$app && bun run build) || exit 1; done

## Database

migrate-up: ## Apply pending migrations
	go run ./cmd/migrate up

migrate-down: ## Roll back the newest migration
	go run ./cmd/migrate down

migrate-status: ## Show applied migrations
	go run ./cmd/migrate status

migrate-new: ## Create a migration — make migrate-new name=add_x
	@test -n "$(name)" || { echo "usage: make migrate-new name=add_x"; exit 1; }
	go tool goose -dir migrations create $(name) go

db-create db-reset db-psql: ## Create, reset or open the database in .env
	@scripts/db.sh $(@:db-%=%)

## Deploy

deploy-build: ## Build the production images
	VERSION=$(VERSION) $(COMPOSE) build

deploy-up: ## Start the production stack in the background
	VERSION=$(VERSION) $(COMPOSE) up -d --build

deploy-down: ## Stop the production stack
	$(COMPOSE) down

deploy-logs: ## Follow the API's logs in the production stack
	$(COMPOSE) logs -f api

deploy-local: ## Start the whole stack on this machine — https://id.localhost
	@VERSION=$(VERSION) scripts/deploy-local.sh

deploy-local-down: ## Stop the local stack and drop its database
	$(COMPOSE_LOCAL) down -v

deploy-local-logs: ## Follow the API's logs in the local stack
	$(COMPOSE_LOCAL) logs -f api

clean: ## Remove build output and logs
	rm -rf bin .logs $(foreach app,$(APPS),web/$(app)/build web/$(app)/.svelte-kit)
