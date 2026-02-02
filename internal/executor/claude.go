package executor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"claude-schedule/internal/db"
)

// baseArgs are the flags shared by every invocation.
var baseArgs = []string{
	"--output-format", "stream-json",
	"--verbose",
	"--dangerously-skip-permissions",
	"--append-system-prompt", "You have access to WebSearch and WebFetch tools. Use them whenever the task requires current or real-time information such as weather, news, prices, or live data. Do not tell the user to check a website themselves - use your tools to fetch the information directly.",
}

// defaultTools are the built-in tools always allowed.
var defaultTools = "Bash,Read,Write,Edit,WebFetch,WebSearch"

// ClaudeExecute runs a job's prompt through the Claude Code CLI and returns the
// response text. It tries to resume the job's previous session for continuity;
// if no session exists yet it falls back to a fresh session.
func ClaudeExecute(ctx context.Context, job db.Job, mcpServers []db.MCPServer) (string, error) {
	mcpArgs, cleanup, err := buildMCPArgs(mcpServers)
	if err != nil {
		return "", fmt.Errorf("building MCP config: %w", err)
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
	out, err := runClaude(ctx, args)
	if err != nil && strings.Contains(err.Error(), "No conversation found") {
		// First run for this job — start a fresh session.
		args = append([]string{"-p", job.Prompt}, allBase...)
		out, err = runClaude(ctx, args)
	}
	return out, err
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
	tb.buf.WriteString(text)
	tb.buf.WriteString("\n\n")
	tb.hasContent = true
}

func (tb *transcriptBuilder) writeToolUse(name string, rawInput json.RawMessage) {
	tb.buf.WriteString("**Tool: " + name + "**\n")
	if len(rawInput) > 0 {
		// Pretty-print the JSON input.
		var parsed interface{}
		if err := json.Unmarshal(rawInput, &parsed); err == nil {
			if pretty, err := json.MarshalIndent(parsed, "", "  "); err == nil {
				tb.buf.WriteString("```json\n" + string(pretty) + "\n```\n\n")
				tb.hasContent = true
				return
			}
		}
		// Fallback: raw JSON.
		tb.buf.WriteString("```json\n" + string(rawInput) + "\n```\n\n")
	} else {
		tb.buf.WriteString("\n")
	}
	tb.hasContent = true
}

func (tb *transcriptBuilder) writeToolResult(text string) {
	if text == "" {
		return
	}
	tb.buf.WriteString("**Result:**\n```\n" + text + "\n```\n\n")
	tb.hasContent = true
}

func (tb *transcriptBuilder) writeSummary(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	tb.buf.WriteString("---\n\n## Summary\n\n" + text + "\n")
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
		existing := tb.build()
		trimmedResult := strings.TrimSpace(lastResult)
		if existing != trimmedResult {
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

// runClaude executes the claude CLI with stream-json output and builds a transcript.
func runClaude(ctx context.Context, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, "claude", args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("creating stdout pipe: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("starting claude: %w", err)
	}

	// Read all lines from stdout.
	var lines []string
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" && len(lines) > 0 {
			errMsg = strings.Join(lines, "\n")
		}
		return "", fmt.Errorf("claude exited with error: %w\n%s", err, errMsg)
	}

	transcript := buildTranscript(lines)
	if transcript == "" {
		raw := strings.Join(lines, "\n")
		if raw == "" {
			return "", fmt.Errorf("empty response from claude")
		}
		return raw, nil
	}

	return transcript, nil
}
