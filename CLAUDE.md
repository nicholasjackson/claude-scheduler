# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Claude Scheduler is a cross-platform tool for scheduling tasks using Claude Code. It supports:
- Configuring MCP servers, skills, and plugins
- Configuring prompts to run at specific times or intervals
- A UI for viewing previous runs with outputs and errors
- Can run on Windows, macOS, and Linux
- Runs in the background with systray icon and native OS notifications
- Claude Scheduler is open source and built with Wails v3, Go and React.

## Build & Development Commands

| Command | Description |
|---------|-------------|
| `make dev` | Run with live reload (Wails dev mode) |
| `make build` | Build production binary to `build/bin/` |
| `make test` | Run all tests (Go + frontend) |
| `make test-go` | Run Go tests only (`go test -v -race ./...`) |
| `make test-frontend` | Run frontend tests only (Vitest) |
| `make lint` | Run all linters (Go + frontend) |
| `make lint-go` | Run Go linter (`go vet`) |
| `make lint-frontend` | Run frontend linter (ESLint) |
| `make generate` | Generate Wails JS/TS bindings |
| `make clean` | Remove build artifacts |
| `make help` | Show all available targets |

## Architecture

**Wails v3** (alpha) desktop app with Go backend and React/TypeScript frontend:
- Go code lives at the repo root (`main.go`, `app.go`)
- `app.go` implements the Wails v3 Service interface (`ServiceStartup`, `ServiceShutdown`)
- React/TypeScript frontend lives in `frontend/` (Vite build system)
- `//go:embed all:frontend/dist` in `main.go` embeds compiled frontend into the Go binary
- Public methods on Go service structs are auto-generated as JS/TS bindings via `wails3 generate bindings`
- Frontend calls Go methods via `Call.ByName()` from `@wailsio/runtime` (see `frontend/src/wailsbridge.ts`)
- Events use `Events.On()` from `@wailsio/runtime` (replaces v2 `EventsOn`)
- Build system uses Taskfile.yml (Task runner) alongside Makefile
- Systray support via `app.SystemTray.New()` — app hides to tray on window close
- Native OS notifications via `github.com/wailsapp/wails/v3/pkg/services/notifications`

## CI/CD

- **CI** (`.github/workflows/ci.yml`): Runs Go tests, frontend lint, and frontend tests on every push/PR to `main`
- **Release** (`.github/workflows/release.yml`): On `v*` tag push, runs tests then builds cross-platform binaries (Linux, Windows, macOS). Automatically uploads artifacts.
- Create a release: `git tag v0.1.0 && git push origin v0.1.0`

## Prerequisites

- Go 1.24+
- Node.js 18+
- Wails CLI v3: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- Task runner: `go install github.com/go-task/task/v3/cmd/task@latest`
- Linux: `sudo apt-get install build-essential libgtk-3-dev libwebkit2gtk-4.1-dev libsoup-3.0-dev`
