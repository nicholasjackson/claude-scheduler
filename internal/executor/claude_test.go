package executor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// assistantLine builds a JSONL line for an assistant event with the given
// content blocks.
func assistantLine(blocks ...cliContentBlock) string {
	evt := cliEvent{
		Type: "assistant",
		Message: &cliMessage{
			Role:    "assistant",
			Content: blocks,
		},
	}
	b, _ := json.Marshal(evt)
	return string(b)
}

// resultLine builds a JSONL line for a result event.
func resultLine(result string) string {
	evt := cliEvent{Type: "result", Result: result}
	b, _ := json.Marshal(evt)
	return string(b)
}

// systemLine builds a JSONL line for a system init event.
func systemLine() string {
	evt := cliEvent{Type: "system", Subtype: "init"}
	b, _ := json.Marshal(evt)
	return string(b)
}

func TestBuildTranscript_TextOnly(t *testing.T) {
	lines := []string{
		systemLine(),
		assistantLine(cliContentBlock{Type: "text", Text: "Hello world"}),
		resultLine("Hello world"),
	}

	got := buildTranscript(lines)
	require.Contains(t, got, "Hello world")
	// Result should NOT duplicate when it matches assistant text.
	require.NotContains(t, got, "## Summary")
}

func TestBuildTranscript_ToolUseWithResult(t *testing.T) {
	lines := []string{
		systemLine(),
		assistantLine(
			cliContentBlock{
				Type:  "tool_use",
				Name:  "Bash",
				ID:    "tool-1",
				Input: json.RawMessage(`{"command":"ls"}`),
			},
		),
		assistantLine(
			cliContentBlock{Type: "text", Text: "Here are the files."},
		),
		resultLine("Here are the files."),
	}

	got := buildTranscript(lines)
	require.Contains(t, got, "**Tool: Bash**")
	require.Contains(t, got, `"command": "ls"`)
	require.Contains(t, got, "Here are the files.")
}

func TestBuildTranscript_MixedTextAndTool(t *testing.T) {
	lines := []string{
		systemLine(),
		assistantLine(
			cliContentBlock{Type: "text", Text: "Let me check."},
			cliContentBlock{
				Type:  "tool_use",
				Name:  "Read",
				Input: json.RawMessage(`{"file":"main.go"}`),
			},
		),
		assistantLine(
			cliContentBlock{Type: "text", Text: "Here is the file."},
		),
		resultLine("Here is the file."),
	}

	got := buildTranscript(lines)
	require.Contains(t, got, "Let me check.")
	require.Contains(t, got, "**Tool: Read**")
	require.Contains(t, got, "Here is the file.")
}

func TestBuildTranscript_EmptyInput(t *testing.T) {
	got := buildTranscript(nil)
	require.Equal(t, "", got)

	got = buildTranscript([]string{})
	require.Equal(t, "", got)
}

func TestBuildTranscript_MalformedLines(t *testing.T) {
	lines := []string{
		"not json at all",
		"",
		`{"type":"unknown"}`,
		assistantLine(cliContentBlock{Type: "text", Text: "valid text"}),
		resultLine("valid text"),
	}

	got := buildTranscript(lines)
	require.Contains(t, got, "valid text")
}

func TestBuildTranscript_ToolWithEmptyInput(t *testing.T) {
	lines := []string{
		assistantLine(cliContentBlock{Type: "tool_use", Name: "WebSearch"}),
	}

	got := buildTranscript(lines)
	require.Contains(t, got, "**Tool: WebSearch**")
	// No json code block since input is empty.
	require.NotContains(t, got, "```json")
}

func TestBuildTranscript_ResultDiffersFromText(t *testing.T) {
	lines := []string{
		assistantLine(
			cliContentBlock{Type: "text", Text: "Let me look into that."},
		),
		resultLine("The answer is 42."),
	}

	got := buildTranscript(lines)
	require.Contains(t, got, "Let me look into that.")
	require.Contains(t, got, "## Summary")
	require.Contains(t, got, "The answer is 42.")
}

func TestBuildTranscript_IgnoresSystemEvents(t *testing.T) {
	lines := []string{
		systemLine(),
		assistantLine(cliContentBlock{Type: "text", Text: "hello"}),
		resultLine("hello"),
	}

	got := buildTranscript(lines)
	require.Equal(t, "hello", got)
	require.NotContains(t, got, "system")
}

func TestBuildMCPArgs_NoServers(t *testing.T) {
	args, cleanup, err := buildMCPArgs(nil)
	require.NoError(t, err)
	require.Nil(t, args)
	cleanup() // noop, should not panic
}
