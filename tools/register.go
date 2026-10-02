package tools

import (
	"context"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// ToolRun is the only tool. Host mcp.toml name is shell → shell__command_run.
const (
	ToolRun      = "command_run"
	descRun      = "Run one shell command in the workspace and return stdout, stderr, and the exit code."
	defaultMS    = int(defaultTimeout / time.Millisecond)
	maxTimeoutMS = int(maxTimeout / time.Millisecond)
)

// ToolNames is the one-tool catalog.
func ToolNames() []string {
	return []string{ToolRun}
}

// Register publishes command_run. root is an already-cleaned jail.
func Register(s *mcpserver.MCPServer, root string) (int, error) {
	s.AddTool(mcp.NewTool(ToolRun,
		mcp.WithDescription(descRun+" Tests, builds, and linters are commands."),
		mcp.WithString("command", mcp.Required(), mcp.Description("One shell command. The working directory is the workspace root.")),
		mcp.WithNumber("timeout_ms", mcp.Description("Timeout in milliseconds. Default 30000, max 120000.")),
		mcp.WithDestructiveHintAnnotation(true),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		command, err := req.RequireString("command")
		if err != nil || strings.TrimSpace(command) == "" {
			return mcp.NewToolResultError("command is required"), nil
		}
		ms := req.GetInt("timeout_ms", defaultMS)
		if ms < 1 {
			ms = defaultMS
		}
		if ms > maxTimeoutMS {
			ms = maxTimeoutMS
		}
		text, err := Run(ctx, root, command, time.Duration(ms)*time.Millisecond)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	})
	return 1, nil
}
