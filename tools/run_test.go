package tools

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/shotah/shell-mcp/server"
)

func TestToolNamesLocked(t *testing.T) {
	t.Parallel()
	re := regexp.MustCompile(`^[a-z]+_[a-z]+`)
	names := ToolNames()
	if len(names) != 1 || names[0] != ToolRun {
		t.Fatalf("names = %v", names)
	}
	if !re.MatchString(names[0]) {
		t.Fatalf("tool %q does not match ^[a-z]+_[a-z]+", names[0])
	}
	if strings.HasPrefix(names[0], server.ServerName) {
		t.Fatalf("tool %q starts with server id %s", names[0], server.ServerName)
	}
}

func TestDescription(t *testing.T) {
	t.Parallel()
	s := server.New()
	if _, err := Register(s, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	resp := s.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	result, ok := resp.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("expected JSONRPCResponse, got %T", resp)
	}
	listResult, ok := result.Result.(mcp.ListToolsResult)
	if !ok {
		t.Fatalf("expected ListToolsResult, got %T", result.Result)
	}
	if len(listResult.Tools) != 1 {
		t.Fatalf("tools = %d", len(listResult.Tools))
	}
	if !strings.HasPrefix(listResult.Tools[0].Description, descRun) {
		t.Fatalf("description = %q", listResult.Tools[0].Description)
	}
}

func TestRunEchoAndFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ctx := context.Background()
	got, err := Run(ctx, root, "echo hello", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "exit: 0\n") || !strings.Contains(got, "hello") {
		t.Fatalf("echo = %q", got)
	}
	got, err = Run(ctx, root, "exit 3", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "exit: 3\n") {
		t.Fatalf("exit = %q", got)
	}
	got, err = Run(ctx, root, "pwd", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, root) {
		t.Fatalf("pwd = %q, want %s", got, root)
	}
	if _, err := Run(ctx, root, "   ", time.Second); err == nil {
		t.Fatal("blank command was accepted")
	}
}

func TestRunTimeoutKeepsTail(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	start := time.Now()
	got, err := Run(context.Background(), root, "echo start; sleep 5; echo end", 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("timeout returned after %s", time.Since(start))
	}
	if !strings.HasPrefix(got, "exit: 124\n") || !strings.Contains(got, "timed out") {
		t.Fatalf("timeout = %q", got)
	}
	if !strings.Contains(got, "start") {
		t.Fatalf("timeout dropped the output: %q", got)
	}
}
