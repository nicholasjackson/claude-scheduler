# Context

## Current Architecture

The scheduler (`internal/scheduler/scheduler.go`) uses a pluggable `ExecuteFunc`
type: `func(ctx context.Context, job db.Job) (string, error)`. Currently `nil`
is passed in `app.go`, which falls back to `mockExecute` (sleeps 30s).

The executor is called from `scheduler.executeJob()` which:
1. Marks the job as "running" and emits a UI event
2. Calls `execFn(ctx, job)` to get output and error
3. Sets status to "success" or "failed" based on the error
4. Stores the output string in `job.Output`
5. Updates timing fields and emits another UI event

## Claude Code CLI

The `claude` CLI supports non-interactive mode via `-p`:
- `claude -p "prompt"` -- runs prompt and exits
- `--session-id <uuid>` -- uses a specific session for context isolation
- `--output-format json` -- returns structured JSON with `result` field
- `--max-turns N` -- limits agentic turns
- `--verbose` -- includes turn-by-turn details in streaming output

JSON output structure:
```json
{
  "result": "The response text...",
  "session_id": "uuid",
  ...
}
```

## Key Files

- `internal/scheduler/scheduler.go` -- Scheduler with ExecuteFunc interface
- `internal/scheduler/scheduler_test.go` -- Tests using mock executors
- `internal/db/jobs.go` -- Job struct with Prompt and Output fields
- `app.go` -- Wires scheduler with store, emitter, and executor
