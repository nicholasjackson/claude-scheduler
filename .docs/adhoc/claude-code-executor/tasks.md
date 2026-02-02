# Tasks

## Phase 1: Create the Claude Code executor
- [ ] Create `internal/executor/claude.go` with `ClaudeExecute` function
- [ ] Implement command building with `-p`, `--session-id`, `--output-format json`, `--dangerously-skip-permissions`
- [ ] Implement JSON response parsing to extract `result` field
- [ ] Handle error cases (non-zero exit, context cancellation, parse failures)

## Phase 2: Wire the executor into the app
- [ ] Import `executor` package in `app.go`
- [ ] Replace `nil` with `executor.ClaudeExecute` in `scheduler.New()` call

## Phase 3: Add tests
- [ ] Create `internal/executor/claude_test.go`
- [ ] Test JSON response parsing (success, malformed, empty)
- [ ] Test error handling (non-zero exit, stderr content)
- [ ] Verify existing scheduler tests still pass
