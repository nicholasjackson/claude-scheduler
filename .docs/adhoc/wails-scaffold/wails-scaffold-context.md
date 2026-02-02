# Wails Project Scaffold — Quick Reference

**Created**: 2026-01-30
**Type**: Ad-hoc

---

## Quick Summary

Scaffold the Claude Scheduler Wails v2 project with Go + React/TypeScript, a Makefile for dev tasks, and GitHub Actions for CI (tests on push/PR) and releases (cross-platform builds on tags).

## Key Files & Locations

### Files to Create
| File | Purpose |
|------|---------|
| `main.go` | Wails entry point, embeds frontend assets |
| `app.go` | Application struct bound to frontend |
| `go.mod` | Go module definition |
| `wails.json` | Wails project configuration |
| `Makefile` | Development task targets |
| `frontend/` | React + TypeScript frontend (Vite) |
| `.github/workflows/ci.yml` | Test on push/PR |
| `.github/workflows/release.yml` | Build + release on tags |

### Files to Modify
| File | Change |
|------|--------|
| `CLAUDE.md` | Add build commands and project structure |
| `.gitignore` | Ensure build artifacts excluded |

## Dependencies

### Go Dependencies
- `github.com/wailsapp/wails/v2` — Wails framework

### Frontend Dependencies (via npm)
- `react`, `react-dom` — UI framework
- `typescript` — Type safety
- `vite`, `@vitejs/plugin-react` — Build tooling

### System Dependencies (Linux)
- `build-essential`, `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`

### CI Dependencies
- `dAppServer/wails-build-action@v3` — Cross-platform Wails builds

## Key Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| CI approach | `dAppServer/wails-build-action` | Community standard, officially referenced |
| Template | Standard `react-ts` | Wails conventions, simple starting point |
| Workflows | Separate CI + Release files | Clearer separation of concerns |
| Release trigger | Tags matching `v*` | Standard semver workflow |

## Integration Points

- **Wails bindings**: Go public methods on bound structs → auto-generated JS in `frontend/wailsjs/go/`
- **Asset embedding**: `//go:embed all:frontend/dist` in `main.go`
- **GitHub Releases**: `wails-build-action` with `package: true` auto-uploads on tag push

## Environment Requirements

- Go 1.23+
- Node.js 18+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Linux: `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`
- macOS: Xcode command line tools
- Windows: WebView2 runtime (usually pre-installed on Windows 10+)
