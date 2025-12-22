package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/mcptools"
)

func main() {
	ctx := context.Background()
	mcpClient, err := mcptools.NewStdioMCPClient(ctx,
		"docker",
		[]string{}, // Environment variables for the MCP client
		"mcp",
		"gateway",
		"run",
	)
	if err != nil {
		panic(err)
	}

	// Print available tools
	for _, tool := range mcpClient.GetTools() {
		println("Tool:", tool.Name, "-", tool.Description)
	}

}
