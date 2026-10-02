package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/shotah/shell-mcp/server"
)

func TestCleanRoot(t *testing.T) {
	t.Parallel()
	if _, err := CleanRoot(""); err == nil {
		t.Fatal("empty root accepted")
	}
	if _, err := CleanRoot(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing root accepted")
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CleanRoot(file); err == nil {
		t.Fatal("file root accepted")
	}
	got, err := CleanRoot(t.TempDir())
	if err != nil || got == "" {
		t.Fatalf("CleanRoot: %v %q", err, got)
	}
}

func TestRunEdges(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ctx := context.Background()
	got, err := Run(ctx, root, "echo err >&2", time.Second)
	if err != nil || !strings.Contains(got, "err") || strings.Contains(got, "stdout") {
		t.Fatalf("stderr = %q %v", got, err)
	}
	got, err = Run(ctx, root, "echo out; echo err >&2", time.Second)
	if err != nil || !strings.Contains(got, "--- stdout ---") || !strings.Contains(got, "--- stderr ---") {
		t.Fatalf("both = %q %v", got, err)
	}
	got, err = Run(ctx, root, "true", 0)
	if err != nil || !strings.HasPrefix(got, "exit: 0\n") {
		t.Fatalf("default timeout = %q %v", got, err)
	}
	got, err = Run(ctx, root, "echo clamped", maxTimeout+time.Second)
	if err != nil || !strings.Contains(got, "clamped") {
		t.Fatalf("clamped = %q %v", got, err)
	}
	big := "printf '%s' '" + strings.Repeat("x", maxBody+80) + "'"
	got, err = Run(ctx, root, big, time.Second)
	if err != nil || !strings.Contains(got, "truncated head") {
		t.Fatalf("tail = %q %v", got, err)
	}
	if _, err := Run(ctx, root, strings.Repeat("a", maxCommand+1), time.Second); err == nil {
		t.Fatal("long command accepted")
	}
	if _, err := Run(ctx, root, "echo\x00hi", time.Second); err == nil {
		t.Fatal("nul command accepted")
	}
	file := filepath.Join(t.TempDir(), "notdir")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(ctx, file, "echo hi", time.Second); err == nil {
		t.Fatal("run in a file succeeded")
	}
}

func TestProcessGroupCancel(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	idle := startShell(context.Background(), "/bin/sh", "true")
	if err := idle.Cancel(); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("cancel before start = %v", err)
	}
	cmd := startShell(context.Background(), "/bin/sh", "true")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Cancel(); err != nil {
		t.Fatal(err)
	}
}

func TestToolHandler(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	s := server.New()
	if _, err := Register(s, root); err != nil {
		t.Fatal(err)
	}
	missing, isErr := callTool(t, s, ToolRun, map[string]any{})
	if !isErr || !strings.Contains(missing, "command") {
		t.Fatalf("missing = %q err=%v", missing, isErr)
	}
	got, isErr := callTool(t, s, ToolRun, map[string]any{"command": "echo hi", "timeout_ms": 0})
	if isErr || !strings.Contains(got, "hi") {
		t.Fatalf("echo = %q err=%v", got, isErr)
	}
	got, isErr = callTool(t, s, ToolRun, map[string]any{"command": "echo capped", "timeout_ms": maxTimeoutMS + 1})
	if isErr || !strings.Contains(got, "capped") {
		t.Fatalf("capped = %q err=%v", got, isErr)
	}
}

func callTool(t *testing.T, s *mcpserver.MCPServer, name string, args map[string]any) (string, bool) {
	t.Helper()
	params := map[string]any{"name": name, "arguments": args}
	msg := map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": params}
	raw, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	resp := s.HandleMessage(context.Background(), raw)
	switch r := resp.(type) {
	case mcp.JSONRPCResponse:
		result, ok := r.Result.(*mcp.CallToolResult)
		if !ok {
			t.Fatalf("result %T", r.Result)
		}
		if len(result.Content) == 0 {
			return "", result.IsError
		}
		text, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatalf("content %T", result.Content[0])
		}
		return text.Text, result.IsError
	case mcp.JSONRPCError:
		t.Fatalf("protocol error %d: %s", r.Error.Code, r.Error.Message)
	default:
		t.Fatalf("response %T", resp)
	}
	return "", true
}
