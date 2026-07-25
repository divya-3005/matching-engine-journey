GO  := go
BIN := bin

.DEFAULT_GOAL := help

.PHONY: help build test race bench fmt vet lint clean run-server run-simulator all

## ── Targets ──────────────────────────────────────────────────────────────────

help: ## Show all available targets
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

build: ## Build server and simulator into bin/
	@mkdir -p $(BIN)
	$(GO) build -o $(BIN)/server    ./cmd/server
	$(GO) build -o $(BIN)/simulator ./cmd/simulator

test: ## Run the test suite
	$(GO) test ./...

race: ## Run the test suite with the race detector
	$(GO) test -race ./...

bench: ## Run benchmarks
	$(GO) test -bench=. ./...

fmt: ## Format all Go source files
	gofmt -w .

vet: ## Run static analysis
	$(GO) vet ./...

lint: fmt vet ## Run fmt and vet

clean: ## Remove build artefacts (bin/)
	rm -rf $(BIN)

run-server: ## Start the HTTP server on :8080
	$(GO) run ./cmd/server

run-simulator: ## Run the order-flow simulator (accepts -orders, -seed, etc.)
	$(GO) run ./cmd/simulator

all: fmt vet test race build ## fmt + vet + test + race + build
