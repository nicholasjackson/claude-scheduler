# Wails Project Scaffold — Implementation Plan

**Created**: 2026-01-30
**Type**: Ad-hoc
**Mode**: Detailed

---

## Overview

Set up the foundational project scaffold for Claude Scheduler: a Wails v2 desktop application using Go (backend) and React + TypeScript (frontend). This includes the Wails project structure, a Makefile for common development tasks, and GitHub Actions CI/CD with tag-based releases.

## Current State Analysis

- Empty repository with only `README.md`, `.gitignore`, `CLAUDE.md`, and `.vscode/settings.json`
- No source code, build system, or CI/CD configuration exists
- Project language/framework decided: Wails v2 + Go + React/TypeScript

## Desired End State

A fully scaffolded Wails v2 project that:
1. Builds and runs locally with `make dev` and `make build`
2. Has Go and frontend test infrastructure ready
3. Includes a Makefile with all common development targets
4. Has GitHub Actions that run tests on push/PR and create cross-platform releases on tags
5. Produces binaries for Linux (amd64), Windows (amd64), and macOS (universal)

## What We're NOT Doing

- Implementing any scheduler business logic
- Building the UI beyond the default template
- Setting up MCP server integration
- Configuring notifications/toast system
- Adding database or persistence layer

## Implementation Approach

Use the standard Wails v2 `react-ts` template as the foundation, then layer on the Makefile and GitHub Actions workflow. The Wails CLI generates the project structure, so we use it directly rather than manually creating files.

---

## Phase 1: Wails Project Initialization

### Overview
Initialize the Wails v2 project using the `react-ts` template. This creates the standard directory structure with Go backend and React/TypeScript frontend.

### Prerequisites
- Go 1.23+ installed
- Node.js 18+ installed
- Wails CLI v2 installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Steps

1. **Install Wails CLI** (if not already installed):
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

2. **Initialize project in current directory:**
   Since the repo already has files, we need to initialize Wails in a temp directory and move files in:
   ```bash
   # Create Wails project in temp dir
   wails init -n claude-schedule -t react-ts -d /tmp/claude-schedule-init

   # Move generated files into the repo (excluding files that already exist)
   # Key files: main.go, app.go, go.mod, go.sum, wails.json, frontend/
   ```

3. **Customize `wails.json`:**
   ```json
   {
     "$schema": "https://wails.io/schemas/config.v2.json",
     "name": "claude-schedule",
     "outputfilename": "claude-schedule",
     "frontend:install": "npm install",
     "frontend:build": "npm run build",
     "frontend:dev:watcher": "npm run dev",
     "frontend:dev:serverUrl": "auto",
     "author": {
       "name": "Nicholas Jackson",
       "email": ""
     }
   }
   ```

4. **Update `main.go`** — Set window title to "Claude Scheduler":
   ```go
   err := wails.Run(&options.App{
       Title:  "Claude Scheduler",
       Width:  1024,
       Height: 768,
       // ... rest of options
   })
   ```

5. **Verify** the scaffold builds and runs:
   ```bash
   wails build
   wails dev  # Should open the app window
   ```

### Success Criteria

#### Automated:
- [ ] `wails build` completes without errors
- [ ] `go test ./...` passes (even if no tests yet)
- [ ] `cd frontend && npm run build` completes without errors

#### Manual:
- [ ] `wails dev` opens a desktop window with the default React template
- [ ] Window title shows "Claude Scheduler"

---

## Phase 2: Makefile

### Overview
Create a Makefile with targets for all common development tasks: building, testing, linting, cleaning, and development mode.

### File: `Makefile` (new)

```makefile
.PHONY: dev build build-debug clean test test-go test-frontend lint lint-go lint-frontend \
        frontend-install generate doctor help

## Run in development mode with live reload
dev:
	wails dev

## Build production binary
build:
	wails build

## Build with debug symbols
build-debug:
	wails build -debug

## Build for specific platforms
build-linux:
	wails build -platform linux/amd64

build-windows:
	wails build -platform windows/amd64

build-darwin:
	wails build -platform darwin/universal

## Clean build artifacts
clean:
	rm -rf build/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules

## Run all tests
test: test-go test-frontend

## Run Go tests
test-go:
	go test -v -race ./...

## Run frontend tests
test-frontend:
	cd frontend && npm test -- --passWithNoTests

## Run all linters
lint: lint-go lint-frontend

## Run Go linter
lint-go:
	go vet ./...

## Run frontend linter
lint-frontend:
	cd frontend && npm run lint

## Install frontend dependencies
frontend-install:
	cd frontend && npm install

## Generate Wails JS bindings
generate:
	wails generate module

## Run Wails system diagnostics
doctor:
	wails doctor

## Show help
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
```

### Notes
- `test-frontend` uses `--passWithNoTests` since the default template has no tests initially
- `lint-go` uses `go vet` (built-in); `golangci-lint` can be added later when more Go code exists
- Platform-specific build targets are provided but CI handles cross-platform builds

### Success Criteria

#### Automated:
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] `make build` produces a binary in `build/bin/`

#### Manual:
- [ ] `make dev` launches the app with live reload
- [ ] `make help` shows all available targets
- [ ] `make clean` removes build artifacts

---

## Phase 3: GitHub Actions CI/CD

### Overview
Set up GitHub Actions with two jobs: (1) test on every push/PR, (2) cross-platform build and release only on pushed tags matching `v*`.

### File: `.github/workflows/ci.yml` (new)

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  NODE_OPTIONS: "--max-old-space-size=4096"

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - uses: actions/setup-node@v4
        with:
          node-version: '18'

      - name: Install Linux build dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y build-essential libgtk-3-dev libwebkit2gtk-4.0-dev

      - name: Run Go tests
        run: go test -v -race ./...

      - name: Install frontend dependencies
        working-directory: frontend
        run: npm ci

      - name: Lint frontend
        working-directory: frontend
        run: npm run lint

      - name: Run frontend tests
        working-directory: frontend
        run: npm test -- --passWithNoTests
```

### File: `.github/workflows/release.yml` (new)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

env:
  NODE_OPTIONS: "--max-old-space-size=4096"

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - uses: actions/setup-node@v4
        with:
          node-version: '18'

      - name: Install Linux build dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y build-essential libgtk-3-dev libwebkit2gtk-4.0-dev

      - name: Run Go tests
        run: go test -v -race ./...

      - name: Install frontend dependencies
        working-directory: frontend
        run: npm ci

      - name: Lint frontend
        working-directory: frontend
        run: npm run lint

      - name: Run frontend tests
        working-directory: frontend
        run: npm test -- --passWithNoTests

  build:
    needs: test
    strategy:
      fail-fast: false
      matrix:
        build:
          - name: claude-schedule
            platform: linux/amd64
            os: ubuntu-latest
          - name: claude-schedule
            platform: windows/amd64
            os: windows-latest
          - name: claude-schedule
            platform: darwin/universal
            os: macos-latest
    runs-on: ${{ matrix.build.os }}
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - uses: dAppServer/wails-build-action@v3
        with:
          build-name: ${{ matrix.build.name }}
          build-platform: ${{ matrix.build.platform }}
          package: true
          go-version: '1.23'
          node-version: '18'
```

### Key Design Decisions

1. **Separate CI and Release workflows**: CI runs on every push/PR to main. Release only runs on tags.
2. **`dAppServer/wails-build-action@v3`**: Handles Go, Node, Wails CLI installation, Linux deps, building, and release upload automatically. When triggered by a tag, it uploads binaries to the GitHub Release.
3. **`package: true`**: The wails-build-action automatically creates/uploads to a GitHub Release when building from a tag push.
4. **Build matrix**: Native runners for each OS (no cross-compilation — macOS builds require macOS runners).
5. **Linux deps**: `libgtk-3-dev` and `libwebkit2gtk-4.0-dev` required for Wails on Ubuntu 22.04 runners.

### Release Workflow

To create a release:
```bash
git tag v0.1.0
git push origin v0.1.0
```

This triggers:
1. Tests run on `ubuntu-latest`
2. If tests pass, three parallel build jobs run (Linux, Windows, macOS)
3. `wails-build-action` with `package: true` automatically uploads binaries to the GitHub Release for the tag

### Success Criteria

#### Automated:
- [ ] Push to `main` triggers CI workflow and tests pass
- [ ] PR to `main` triggers CI workflow and tests pass
- [ ] Pushing a `v*` tag triggers release workflow
- [ ] Release workflow produces binaries for all three platforms

#### Manual:
- [ ] GitHub Release page shows uploaded binaries for Linux, Windows, macOS
- [ ] Downloaded binaries run on their respective platforms

---

## Phase 4: Update CLAUDE.md

### Overview
Update CLAUDE.md with build commands and project structure now that the scaffold exists.

### File: `CLAUDE.md` (update)

Add sections for:
- Build commands (`make dev`, `make build`, `make test`, `make lint`)
- Project structure overview (Go backend in root, React frontend in `frontend/`)
- Architecture: Wails binds Go structs to frontend via auto-generated `frontend/wailsjs/` bindings
- CI/CD: Tests on push/PR, releases on tags

### Success Criteria

#### Automated:
- [ ] CLAUDE.md contains accurate build commands

#### Manual:
- [ ] Future Claude Code instances can understand the project from CLAUDE.md

---

## Testing Strategy

### Go Tests
- Run with `make test-go` or `go test -v -race ./...`
- Uses `testify/require` for assertions (following Go dev guidelines)
- No tests initially — scaffold provides the infrastructure

### Frontend Tests
- Run with `make test-frontend` or `cd frontend && npm test`
- Uses `--passWithNoTests` initially since default template has no tests
- Vitest or Jest can be configured when tests are added

### CI Tests
- Both Go and frontend tests run in CI on every push/PR
- Tests must pass before release builds proceed

## Performance Considerations

- Wails embeds frontend assets into the Go binary (`//go:embed`) — single-file distribution
- Production builds use `-ldflags "-w -s"` to strip debug info
- `NODE_OPTIONS: "--max-old-space-size=4096"` prevents OOM during frontend builds in CI

## References

- [Wails v2 Documentation](https://wails.io/docs/)
- [Wails Cross-Platform Build Guide](https://wails.io/docs/guides/crossplatform-build/)
- [dAppServer/wails-build-action](https://github.com/dAppServer/wails-build-action)
- [Wails react-ts template](https://github.com/wailsapp/wails/tree/master/v2/pkg/templates/templates/react-ts)
