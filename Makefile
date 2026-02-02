.PHONY: dev build build-debug clean test test-go test-frontend lint lint-go lint-frontend \
        build-linux build-windows build-darwin frontend-install frontend-build generate doctor ci help

dev: ## Run in development mode with live reload
	wails3 dev -config ./build/config.yml

build: ## Build production binary for current OS
	task build PRODUCTION=true

build-debug: ## Build with debug symbols for current OS
	task build

build-linux: ## Build for Linux amd64
	task linux:build PRODUCTION=true

build-windows: ## Build for Windows amd64
	task windows:build PRODUCTION=true

build-darwin: ## Build for macOS
	task darwin:build PRODUCTION=true

clean: ## Clean build artifacts
	rm -rf build/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules

test: test-go test-frontend ## Run all tests

test-go: ## Run Go tests
	go test -v -race ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test -- --passWithNoTests

lint: lint-go lint-frontend ## Run all linters

lint-go: ## Run Go linter
	go vet ./...

lint-frontend: ## Run frontend linter
	cd frontend && npm run lint

frontend-install: ## Install frontend dependencies
	cd frontend && npm ci

frontend-build: frontend-install ## Build frontend (needed before Go tests)
	cd frontend && npm run build

ci: frontend-build test lint ## Run full CI checks locally (build frontend, test, lint)

generate: ## Generate Wails JS bindings
	wails3 generate bindings

doctor: ## Run Wails system diagnostics
	wails3 doctor

help: ## Show help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
