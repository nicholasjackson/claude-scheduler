# Wails Project Scaffold — Research Notes

**Created**: 2026-01-30
**Type**: Ad-hoc

---

## Initial Understanding

Set up the foundational scaffold for Claude Scheduler: a Wails v2 desktop application with Go backend and React/TypeScript frontend. Includes Makefile for dev tasks and GitHub Actions for CI/CD with tag-based releases.

## Research Process

### 1. CLAUDE.md Review
- Project: Claude Scheduler — cross-platform task scheduling tool using Claude Code
- Tech stack: Wails, Go, React
- Features: MCP server config, scheduled prompts, run history UI, background notifications
- Platforms: Windows, macOS, Linux

### 2. Wails v2 Project Structure Research

**Source**: Official Wails docs, community examples, template source

**Findings:**
- `wails init -n <name> -t react-ts` generates the standard scaffold
- Key files: `main.go` (entry point with `//go:embed`), `app.go` (bound struct), `wails.json` (config)
- Frontend: Vite + React + TypeScript in `frontend/`
- Auto-generated bindings: `frontend/wailsjs/go/` — JS functions calling Go methods
- Build: `wails build` produces single binary with embedded frontend
- Dev: `wails dev` runs live-reload with Vite dev server

**Key config (`wails.json`):**
- `frontend:install` → `npm install`
- `frontend:build` → `npm run build`
- `frontend:dev:watcher` → `npm run dev`
- `frontend:dev:serverUrl` → `auto` (Wails auto-discovers Vite port)

### 3. GitHub Actions Research

**Source**: Official Wails cross-platform build guide, dAppServer/wails-build-action

**Findings:**
- `dAppServer/wails-build-action@v3` is the community-standard action, referenced in official Wails docs
- Handles: Go setup, Node setup, Wails CLI install, Linux deps, build, artifact upload
- With `package: true` on a tag build, automatically uploads to GitHub Release
- Cross-compilation not recommended — use native runners per platform
- Linux requires: `build-essential libgtk-3-dev libwebkit2gtk-4.0-dev` (Ubuntu 22.04)
- Ubuntu 24.04 needs `libwebkit2gtk-4.1-dev` + `-tags webkit2_41` (future consideration)

**Chosen approach**: `dAppServer/wails-build-action` for builds, manual setup for test job (more control over test steps).

### 4. Go Development Guidelines Review

**Source**: go-dev-guidelines skill

**Key patterns to follow:**
- TDD with `testify/require` (never use standard testing package directly)
- `require` not `assert` (stops on failure)
- Explicit test functions per scenario, no table-driven tests
- Error wrapping with `%w`, sentinel errors with `Err` prefix
- Constructor pattern for DI, no package-level globals
- Three-layer architecture: Handler → Service → Repository

### 5. Past Learnings Search

No learnings directory found — this is a fresh project with no prior institutional knowledge.

## Design Decisions

### Decision 1: Separate CI and Release Workflows
- **Options**: Single workflow with conditional jobs vs. two separate workflows
- **Chosen**: Two separate workflows (`ci.yml` and `release.yml`)
- **Rationale**: Clearer separation of concerns. CI is simple (test only). Release is complex (test + build matrix + publish). Easier to maintain and debug independently.

### Decision 2: wails-build-action for Builds
- **Options**: (A) wails-build-action for everything, (B) Manual setup, (C) Both
- **Chosen**: (A) wails-build-action — user selected this
- **Rationale**: Community standard, officially referenced, handles platform deps automatically, auto-uploads to releases on tags.

### Decision 3: Standard react-ts Template
- **Options**: Standard template vs. custom structure (cmd/, internal/)
- **Chosen**: Standard template — user selected this
- **Rationale**: Wails expects root-level `main.go` and `app.go`. Custom structure can be introduced later as the project grows.

### Decision 4: Wails Init Strategy
- **Challenge**: Repo already has files (README, CLAUDE.md, .gitignore, .vscode/)
- **Approach**: Init in temp dir, move files into repo
- **Rationale**: `wails init` expects an empty target directory. Moving files preserves existing repo content.

## Open Questions

None — all questions resolved during planning.
