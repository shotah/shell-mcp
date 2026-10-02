package main

import (
	"fmt"
	"log"
	"os"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/shotah/shell-mcp/server"
	"github.com/shotah/shell-mcp/tools"
	"github.com/spf13/cobra"
)

func main() {
	var rootPath string
	cmd := &cobra.Command{
		Use:   "shell-mcp",
		Short: "One workspace shell command over MCP stdio",
		RunE: func(_ *cobra.Command, _ []string) error {
			root, err := tools.CleanRoot(rootPath)
			if err != nil {
				return err
			}
			s := server.New()
			n, err := tools.Register(s, root)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "shell-mcp: published %d tools\n", n)
			return mcpserver.ServeStdio(s, mcpserver.WithErrorLogger(log.New(os.Stderr, "", log.LstdFlags)))
		},
	}
	cmd.Flags().StringVar(&rootPath, "root", "", "workspace jail (required)")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
