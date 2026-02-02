# Research

## Claude Code CLI Flags (relevant subset)

| Flag | Description |
|------|-------------|
| `-p` / `--print` | Non-interactive mode: process prompt, print, exit |
| `--session-id <uuid>` | Use specific session ID for context isolation |
| `--output-format json` | Structured JSON output with `result` field |
| `--output-format stream-json` | Newline-delimited JSON for real-time streaming |
| `--max-turns N` | Limit number of agentic turns |
| `--max-budget-usd N` | Cap API spending per invocation |
| `--allowedTools "Bash,Read,Edit"` | Restrict available tools |
| `--dangerously-skip-permissions` | Skip all permission prompts |
| `--continue` | Continue most recent conversation |
| `--resume <id>` | Resume specific session |
| `--verbose` | Turn-by-turn logging |
| `--no-session-persistence` | Don't save session to disk |
| `--append-system-prompt` | Add to system prompt |

## Session Isolation Strategy

Using `--session-id <job.ID>` per job means:
- Each job has its own conversation history
- Repeated runs of the same job accumulate context (useful for iterative tasks)
- Different jobs never share context
- Job IDs are UUIDs, which is the required format for session IDs

## Go exec.CommandContext

```go
cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--session-id", jobID, "--output-format", "json")
cmd.Dir = workingDir
output, err := cmd.Output()  // stdout only; err includes stderr via ExitError
```

- `CommandContext` kills the process when the context is cancelled
- `cmd.Output()` returns stdout; stderr is available via `(*exec.ExitError).Stderr`
- For combined output, use `cmd.CombinedOutput()` instead

## JSON Response Parsing

Only need to extract the `result` field:
```go
type claudeResponse struct {
    Result    string `json:"result"`
    SessionID string `json:"session_id"`
}
```
