package executor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"claude-schedule/internal/db"
)

// DebugDir, when non-empty, causes each CLI invocation's raw JSONL output
// to be written to a timestamped file in that directory.
var DebugDir string

// baseArgs are the flags shared by every invocation.
var baseArgs = []string{
	"--output-format", "stream-json",
	"--verbose",
	"--dangerously-skip-permissions",
	"--append-system-prompt", "You have access to WebSearch and WebFetch tools. Use them whenever the task requires current or real-time information such as weather, news, prices, or live data. Do not tell the user to check a website themselves - use your tools to fetch the information directly.",
}

// defaultTools are the built-in tools always allowed.
var defaultTools = "Bash,Read,Write,Edit,WebFetch,WebSearch"

// ExecuteResult holds the output and raw JSONL lines from a CLI invocation.
type ExecuteResult struct {
	Transcript string
	RawLines   []string
}

// ClaudeExecute runs a job's prompt through the Claude Code CLI and returns the
// response text. It tries to resume the job's previous session for continuity;
// if no session exists yet it falls back to a fresh session.
func ClaudeExecute(ctx context.Context, job db.Job, mcpServers []db.MCPServer) (ExecuteResult, error) {
	mcpArgs, cleanup, err := buildMCPArgs(mcpServers)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("building MCP config: %w", err)
	}
	defer cleanup()

	// Build allowed tools list.
	tools := defaultTools
	for _, srv := range mcpServers {
		tools += ",mcp__" + srv.Name + "__*"
	}

	allBase := append([]string{}, baseArgs...)
	allBase = append(allBase, "--allowedTools", tools)
	allBase = append(allBase, mcpArgs...)

	// Try resuming the previous session first.
	args := append([]string{"-p", job.Prompt, "--resume", job.ID}, allBase...)
	result, err := runClaude(ctx, args)
	if err != nil && strings.Contains(err.Error(), "No conversation found") {
		// First run for this job — start a fresh session.
		args = append([]string{"-p", job.Prompt}, allBase...)
		result, err = runClaude(ctx, args)
	}
	return result, err
}

// ClaudeAnswer resumes a conversation with the user's answer to a question.
func ClaudeAnswer(ctx context.Context, job db.Job, mcpServers []db.MCPServer, answer string) (ExecuteResult, error) {
	mcpArgs, cleanup, err := buildMCPArgs(mcpServers)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("building MCP config: %w", err)
	}
	defer cleanup()

	tools := defaultTools
	for _, srv := range mcpServers {
		tools += ",mcp__" + srv.Name + "__*"
	}

	allBase := append([]string{}, baseArgs...)
	allBase = append(allBase, "--allowedTools", tools)
	allBase = append(allBase, mcpArgs...)

	args := append([]string{"-p", answer, "--resume", job.ID}, allBase...)
	return runClaude(ctx, args)
}

// DetectQuestion scans raw JSONL lines for the last AskUserQuestion tool call
// and returns the question JSON string (or empty if none found).
func DetectQuestion(lines []string) string {
	var lastQuestion string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var evt cliEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue
		}
		if evt.Type != "assistant" || evt.Message == nil {
			continue
		}
		for _, block := range evt.Message.Content {
			if block.Type == "tool_use" && block.Name == "AskUserQuestion" && len(block.Input) > 0 {
				var qi questionInput
				if err := json.Unmarshal(block.Input, &qi); err == nil && len(qi.Questions) > 0 {
					lastQuestion = string(block.Input)
				}
			}
		}
	}
	return lastQuestion
}

// ---------------------------------------------------------------------------
// CLI stream-json event types
// ---------------------------------------------------------------------------

// cliEvent is the top-level envelope for every JSONL line emitted by
// `claude -p --output-format stream-json`.
type cliEvent struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype,omitempty"`
	Message *cliMessage     `json:"message,omitempty"`  // present when Type == "assistant"
	Content json.RawMessage `json:"content,omitempty"`  // present for tool result events
	Result  string          `json:"result,omitempty"`   // present when Type == "result"
	IsError bool            `json:"is_error,omitempty"` // true when Type == "result" and the run failed
}

// cliMessage mirrors the Anthropic API Message structure embedded in
// assistant-type events.
type cliMessage struct {
	Role    string            `json:"role"`
	Content []cliContentBlock `json:"content"`
}

// cliContentBlock represents one block inside an assistant message.
type cliContentBlock struct {
	Type  string          `json:"type"`            // "text", "tool_use", "tool_result"
	Text  string          `json:"text,omitempty"`  // for type == "text"
	Name  string          `json:"name,omitempty"`  // for type == "tool_use"
	ID    string          `json:"id,omitempty"`    // for type == "tool_use"
	Input json.RawMessage `json:"input,omitempty"` // for type == "tool_use"
}

// ---------------------------------------------------------------------------
// Transcript builder
// ---------------------------------------------------------------------------

// transcriptBuilder accumulates CLI events into a markdown transcript.
type transcriptBuilder struct {
	buf        strings.Builder
	hasContent bool
}

func (tb *transcriptBuilder) writeText(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	tb.buf.WriteString(`<div style="margin:8px 0;padding:8px 12px;border-left:3px solid #22d3ee;background:#0f172a;border-radius:4px;color:#e2e8f0;font-size:13px;line-height:1.5">`)
	tb.buf.WriteString(text)
	tb.buf.WriteString("</div>\n\n")
	tb.hasContent = true
}

// questionInput matches the AskUserQuestion tool input shape.
type questionInput struct {
	Questions []questionItem `json:"questions"`
}

type questionItem struct {
	Question string           `json:"question"`
	Header   string           `json:"header"`
	Options  []questionOption `json:"options"`
}

type questionOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

func (tb *transcriptBuilder) writeToolUse(name string, rawInput json.RawMessage) {
	// Render AskUserQuestion tool calls as a styled question card.
	if name == "AskUserQuestion" && len(rawInput) > 0 {
		var qi questionInput
		if err := json.Unmarshal(rawInput, &qi); err == nil && len(qi.Questions) > 0 {
			tb.writeQuestion(qi)
			return
		}
	}

	tb.buf.WriteString(`<details style="margin:8px 0;border:1px solid #374151;border-radius:6px;overflow:hidden">`)
	tb.buf.WriteString(`<summary style="cursor:pointer;padding:6px 10px;background:#1e293b;color:#60a5fa;font-size:13px;font-weight:600">`)
	tb.buf.WriteString(`Tool: ` + name)
	tb.buf.WriteString(`</summary>`)

	if len(rawInput) > 0 {
		input := string(rawInput)
		var parsed interface{}
		if err := json.Unmarshal(rawInput, &parsed); err == nil {
			if pretty, err := json.MarshalIndent(parsed, "", "  "); err == nil {
				input = string(pretty)
			}
		}
		tb.buf.WriteString(`<pre style="margin:0;padding:8px 10px;background:#0f172a;color:#94a3b8;font-size:12px;overflow-x:auto">`)
		tb.buf.WriteString(input)
		tb.buf.WriteString(`</pre>`)
	}

	tb.buf.WriteString("</details>\n\n")
	tb.hasContent = true
}

func (tb *transcriptBuilder) writeQuestion(qi questionInput) {
	for _, q := range qi.Questions {
		tb.buf.WriteString(`<div style="margin:8px 0;padding:12px 16px;border:1px solid #f59e0b;border-radius:6px;background:#1c1917;color:#e2e8f0;font-size:13px;line-height:1.5">`)
		tb.buf.WriteString(`<div style="color:#f59e0b;font-weight:700;font-size:11px;text-transform:uppercase;letter-spacing:0.05em;margin-bottom:6px">`)
		if q.Header != "" {
			tb.buf.WriteString(q.Header)
		} else {
			tb.buf.WriteString("Question")
		}
		tb.buf.WriteString(`</div>`)
		tb.buf.WriteString(`<div style="margin-bottom:10px;font-size:14px">` + q.Question + `</div>`)
		if len(q.Options) > 0 {
			for _, opt := range q.Options {
				tb.buf.WriteString(`<div style="margin:4px 0;padding:6px 10px;border:1px solid #374151;border-radius:4px;background:#0f172a">`)
				tb.buf.WriteString(`<span style="color:#fbbf24;font-weight:600">` + opt.Label + `</span>`)
				if opt.Description != "" {
					tb.buf.WriteString(` <span style="color:#94a3b8;font-size:12px">— ` + opt.Description + `</span>`)
				}
				tb.buf.WriteString(`</div>`)
			}
		}
		tb.buf.WriteString(`</div>`)
		tb.buf.WriteString("\n\n")
	}
	tb.hasContent = true
}

func (tb *transcriptBuilder) writeToolResult(text string) {
	if text == "" {
		return
	}
	tb.buf.WriteString(`<details style="margin:8px 0;border:1px solid #374151;border-radius:6px;overflow:hidden">`)
	tb.buf.WriteString(`<summary style="cursor:pointer;padding:6px 10px;background:#1e293b;color:#a78bfa;font-size:13px;font-weight:600">`)
	tb.buf.WriteString(`Result`)
	tb.buf.WriteString(`</summary>`)
	tb.buf.WriteString(`<pre style="margin:0;padding:8px 10px;background:#0f172a;color:#94a3b8;font-size:12px;overflow-x:auto;white-space:pre-wrap">`)
	tb.buf.WriteString(text)
	tb.buf.WriteString(`</pre>`)
	tb.buf.WriteString("</details>\n\n")
	tb.hasContent = true
}

func (tb *transcriptBuilder) writeSummary(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	tb.buf.WriteString(`<hr style="border-color:#374151;margin:16px 0">`)
	tb.buf.WriteString(`<div style="margin:8px 0;padding:10px 12px;border-left:3px solid #34d399;background:#0f172a;border-radius:4px;color:#e2e8f0;font-size:13px;line-height:1.5">`)
	tb.buf.WriteString(`<strong style="color:#34d399">Summary</strong><br>`)
	tb.buf.WriteString(text)
	tb.buf.WriteString("</div>\n")
	tb.hasContent = true
}

// handleAssistant processes an assistant-type event, extracting text and
// tool-use blocks from the message content.
func (tb *transcriptBuilder) handleAssistant(msg *cliMessage) {
	if msg == nil {
		return
	}
	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			tb.writeText(block.Text)
		case "tool_use":
			tb.writeToolUse(block.Name, block.Input)
		case "tool_result":
			tb.writeToolResult(block.Text)
		}
	}
}

func (tb *transcriptBuilder) build() string {
	return strings.TrimSpace(tb.buf.String())
}

// buildTranscript parses JSONL lines from stream-json output and returns a
// markdown transcript showing the thought process, tool calls, and a final
// summary.
func buildTranscript(lines []string) string {
	tb := &transcriptBuilder{}
	var lastResult string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var evt cliEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue
		}

		switch evt.Type {
		case "assistant":
			tb.handleAssistant(evt.Message)
		case "result":
			lastResult = evt.Result
		// "system" and other types are ignored.
		}
	}

	// Append summary from the result event if it differs from what we
	// already captured (avoids duplication when the result just echoes
	// the last assistant text).
	if lastResult != "" {
		trimmedResult := strings.TrimSpace(lastResult)
		if !strings.Contains(tb.build(), trimmedResult) {
			tb.writeSummary(lastResult)
		}
	}

	return tb.build()
}

// mcpConfigFile represents the JSON structure expected by --mcp-config.
type mcpConfigFile struct {
	MCPServers map[string]mcpServerEntry `json:"mcpServers"`
}

type mcpServerEntry struct {
	Type    string            `json:"type"`
	URL     string            `json:"url,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// buildMCPArgs creates a temp MCP config file and returns the CLI args to use it.
// The returned cleanup function removes the temp file.
func buildMCPArgs(servers []db.MCPServer) (args []string, cleanup func(), err error) {
	noop := func() {}
	if len(servers) == 0 {
		return nil, noop, nil
	}

	config := mcpConfigFile{
		MCPServers: make(map[string]mcpServerEntry, len(servers)),
	}

	for _, srv := range servers {
		entry := mcpServerEntry{
			Type: srv.Type,
		}

		if srv.Type == "http" {
			entry.URL = srv.URL
		} else {
			entry.Command = srv.Command
		}

		if srv.Args != "" && srv.Args != "[]" {
			var parsed []string
			if err := json.Unmarshal([]byte(srv.Args), &parsed); err == nil {
				entry.Args = parsed
			}
		}

		if srv.Env != "" && srv.Env != "{}" {
			var parsed map[string]string
			if err := json.Unmarshal([]byte(srv.Env), &parsed); err == nil {
				entry.Env = parsed
			}
		}

		if srv.Headers != "" && srv.Headers != "{}" {
			var parsed map[string]string
			if err := json.Unmarshal([]byte(srv.Headers), &parsed); err == nil {
				entry.Headers = parsed
			}
		}

		config.MCPServers[srv.Name] = entry
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, noop, fmt.Errorf("marshalling MCP config: %w", err)
	}

	f, err := os.CreateTemp("", "claude-mcp-*.json")
	if err != nil {
		return nil, noop, fmt.Errorf("creating MCP temp file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, noop, fmt.Errorf("writing MCP config: %w", err)
	}
	f.Close()

	cleanup = func() { os.Remove(f.Name()) }
	return []string{"--mcp-config", f.Name()}, cleanup, nil
}

// extractError inspects stream-json output lines for a human-readable error
// message. It prefers the result event (with is_error=true) and falls back to
// assistant message text.
func extractError(lines []string) string {
	var fallback string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var evt cliEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue
		}
		// The result event carries the best summary on error.
		if evt.Type == "result" && evt.IsError && evt.Result != "" {
			return evt.Result
		}
		// Capture assistant text as a fallback.
		if fallback == "" && evt.Type == "assistant" && evt.Message != nil {
			for _, block := range evt.Message.Content {
				if block.Type == "text" && block.Text != "" {
					fallback = block.Text
					break
				}
			}
		}
	}
	return fallback
}

// runClaude executes the claude CLI with stream-json output and builds a transcript.
func runClaude(ctx context.Context, args []string) (ExecuteResult, error) {
	cmd := exec.CommandContext(ctx, "claude", args...)
	hideWindow(cmd)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("creating stdout pipe: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return ExecuteResult{}, fmt.Errorf("starting claude: %w", err)
	}

	// Read all lines from stdout.
	var lines []string
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	dumpDebugLines(lines)

	if err := cmd.Wait(); err != nil {
		// Try to extract a human-readable error from the stream-json output.
		if msg := extractError(lines); msg != "" {
			return ExecuteResult{}, fmt.Errorf("%s", msg)
		}
		// Build the most informative error we can from what's available.
		stderrMsg := strings.TrimSpace(stderr.String())
		stdout := strings.TrimSpace(strings.Join(lines, "\n"))

		var parts []string
		if stderrMsg != "" {
			parts = append(parts, stderrMsg)
		}
		if stdout != "" {
			parts = append(parts, stdout)
		}
		if len(parts) == 0 {
			parts = append(parts, err.Error())
		}
		return ExecuteResult{}, fmt.Errorf("claude: %s", strings.Join(parts, "\n"))
	}

	transcript := buildTranscript(lines)
	if transcript == "" {
		raw := strings.Join(lines, "\n")
		if raw == "" {
			return ExecuteResult{}, fmt.Errorf("empty response from claude")
		}
		return ExecuteResult{Transcript: raw, RawLines: lines}, nil
	}

	return ExecuteResult{Transcript: transcript, RawLines: lines}, nil
}

// dumpDebugLines writes raw JSONL lines to a timestamped file in DebugDir.
func dumpDebugLines(lines []string) {
	if DebugDir == "" || len(lines) == 0 {
		return
	}
	if err := os.MkdirAll(DebugDir, 0o755); err != nil {
		log.Printf("debug: failed to create dir %s: %v", DebugDir, err)
		return
	}
	name := fmt.Sprintf("run-%s.jsonl", time.Now().UTC().Format("20060102-150405"))
	path := filepath.Join(DebugDir, name)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		log.Printf("debug: failed to write %s: %v", path, err)
	} else {
		log.Printf("debug: wrote %d lines to %s", len(lines), path)
	}
}
