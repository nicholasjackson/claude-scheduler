# Wails Project Scaffold — Task Checklist

**Created**: 2026-01-30
**Type**: Ad-hoc

---

## Phase 1: Wails Project Initialization

- [ ] Install Wails CLI if not present (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) — `S`
- [ ] Run `wails init -n claude-schedule -t react-ts` in temp directory — `S`
- [ ] Move generated files into repo root (preserving existing files) — `S`
- [ ] Update `wails.json` with project name and author — `S`
- [ ] Update `main.go` window title to "Claude Scheduler" — `S`
- [ ] Update `.gitignore` to exclude build artifacts — `S`
- [ ] Verify: `wails build` succeeds — `S`
- [ ] Verify: `go test ./...` passes — `S`
- [ ] Verify: `cd frontend && npm run build` succeeds — `S`

## Phase 2: Makefile

- [ ] Create `Makefile` with targets: dev, build, build-debug, build-{linux,windows,darwin}, clean, test, test-go, test-frontend, lint, lint-go, lint-frontend, frontend-install, generate, doctor, help — `M`
- [ ] Verify: `make test` passes — `S`
- [ ] Verify: `make lint` passes — `S`
- [ ] Verify: `make build` produces binary in `build/bin/` — `S`

## Phase 3: GitHub Actions CI/CD

- [ ] Create `.github/workflows/ci.yml` — test on push/PR to main — `M`
- [ ] Create `.github/workflows/release.yml` — test + build + release on `v*` tags — `M`
  - Uses `dAppServer/wails-build-action@v3` with build matrix (linux, windows, macos)
  - `package: true` for auto-upload to GitHub Release
- [ ] Verify: Push to main triggers CI workflow — `S`

## Phase 4: Update CLAUDE.md

- [ ] Add build/dev/test/lint commands to CLAUDE.md — `S`
- [ ] Add project structure overview — `S`
- [ ] Add architecture notes (Wails binding, asset embedding) — `S`

---

## Final Verification

### Automated:
- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] `wails dev` launches without errors

### Manual:
- [ ] App window opens with "Claude Scheduler" title
- [ ] Default React template renders in the window
- [ ] `make help` shows all targets
- [ ] GitHub Actions workflows are syntactically valid (push to verify)

---

## Notes

- Effort estimates: S = Small, M = Medium, L = Large
- Phase 1 depends on Wails CLI being installed
- Phase 3 can only be fully verified after pushing to GitHub
- Ubuntu 24.04 runners may need `libwebkit2gtk-4.1-dev` + `-tags webkit2_41` in the future
