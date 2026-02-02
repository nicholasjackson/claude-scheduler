# Plan: Execute Jobs via Claude Code CLI

## Summary

Replace the mock 30-second sleep executor with a real implementation that runs
each job's prompt through the `claude` CLI in non-interactive mode (`-p`). Each
job gets its own session ID so context stays isolated. Output is captured and
stored in the job's `output` field.

## Design Decisions

1. **Session isolation** -- Each job uses its own deterministic session ID
   derived from the job's UUID (`--session-id <job.ID>`). This means repeated
   runs of the same job share conversational context, while different jobs are
   fully isolated.

2. **JSON output** -- Use `--output-format json` so we can reliably parse the
   result text from the structured response rather than scraping stdout.

3. **Context cancellation** -- The `ExecuteFunc` already receives a
   `context.Context`. We use `exec.CommandContext` so that stopping the
   scheduler (app shutdown) kills the child process.

4. **Skip permissions** -- Pass `--dangerously-skip-permissions` so that
   scheduled jobs can run unattended without blocking on permission prompts.
   Since the user explicitly configures each job's prompt, this is an
   intentional delegation of trust.

5. **Keep executor pluggable** -- The new executor is a standalone function
   matching the existing `ExecuteFunc` signature. Tests can still inject a fast
   mock. The real executor is wired in `app.go` at startup.

6. **Working directory** -- Jobs run in the user's home directory by default.
   A future enhancement could add a per-job working directory field.

## Phases

### Phase 1: Create the Claude Code executor

Create `internal/executor/claude.go` with a function matching the
`scheduler.ExecuteFunc` signature:

```go
func ClaudeExecute(ctx context.Context, job db.Job) (string, error)
```

Implementation:
- Build the command: `claude -p <job.Prompt> --session-id <job.ID> --output-format json --dangerously-skip-permissions --verbose`
- Use `exec.CommandContext(ctx, ...)` for cancellation support
- Capture combined stdout into a buffer
- Parse the JSON response to extract the `result` field
- Return the result text (or the raw output if JSON parsing fails)
- Return any non-zero exit code as an error, including stderr content

### Phase 2: Wire the executor into the app

In `app.go`, replace the `nil` executor with the new `ClaudeExecute` function:

```go
a.sched = scheduler.New(a.store, emit, executor.ClaudeExecute, 60*time.Second)
```

### Phase 3: Add tests

Create `internal/executor/claude_test.go`:
- Test command construction (verify args include `-p`, `--session-id`, `--output-format json`)
- Test JSON response parsing (valid response, malformed response, empty result)
- Test error handling (non-zero exit code, context cancellation)
- Use a mock/stub for the actual CLI binary (don't call real `claude` in tests)

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `internal/executor/claude.go` | Create | Claude Code CLI executor function |
| `internal/executor/claude_test.go` | Create | Tests for the executor |
| `app.go` | Modify | Wire `executor.ClaudeExecute` instead of `nil` |

## Success Criteria

- Creating a job with a prompt and letting it become due results in `claude -p`
  being invoked with that prompt
- The output shown in the UI is the actual Claude response text
- Each job uses a separate session (verified by `--session-id` matching job ID)
- Stopping the app kills any in-flight `claude` process
- Existing scheduler tests continue to pass (they use their own mock executor)
