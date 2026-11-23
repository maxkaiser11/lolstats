ENV_FILE ?= .env
OP       := op run --env-file=$(ENV_FILE) --
BIN      := bin/lolstats

.DEFAULT_GOAL := help
.PHONY: help run build test test-race fmt vet tidy check clean \
        migrate-up migrate-down migrate-status migrate-reset migrate-create

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## --- App -------------------------------------------------------------

run: ## Run the app with secrets from 1Password
	$(OP) go run ./cmd/api/

build: ## Build binary to bin/
	go build -o $(BIN) ./cmd/api/

clean: ## Remove build output
	rm -rf bin

## --- Quality ---------------------------------------------------------

test: ## Run tests (secrets injected)
	$(OP) go test ./...

test-race: ## Run tests with race detector
	$(OP) go test -race ./...

fmt: ## Format code
	go fmt ./...

vet: ## Vet code
	go vet ./...

tidy: ## Tidy go.mod
	go mod tidy

check: fmt vet test ## fmt + vet + test

## --- Migrations (goose reads GOOSE_* env vars) -----------------------

migrate-up: ## Apply all migrations
	$(OP) goose up

migrate-down: ## Roll back the last migration
	$(OP) goose down

migrate-status: ## Show migration status
	$(OP) goose status

migrate-reset: ## Roll back ALL migrations
	$(OP) goose reset

migrate-create: ## New SQL migration: make migrate-create name=add_matches
	@test -n "$(name)" || (echo "usage: make migrate-create name=<name>"; exit 1)
	$(OP) goose create $(name) sql
