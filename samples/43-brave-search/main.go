package main

import (
	"context"
	"fmt"
	"strings"
	"github.com/snipwise/nova/nova-sdk/mcptools"

)

/*
To execute this sample, make sure you have an MCP server running locally.

```
cd /samples/mcp-servers
docker compose up --build
```
*/

func main() {
	ctx := context.Background()

	mcpClient, err := mcptools.NewStreamableHttpMCPClient(ctx, "http://localhost:9011")

	if err != nil {
		panic(err)
	}

	// Print available tools
	for _, tool := range mcpClient.GetTools() {
		println("Tool:", tool.Name, "-", tool.Description)
	}

	fmt.Println(strings.Repeat("=", 50))

	type BraveInput struct {
		Query string  `json:"query"`
		Count int     `json:"count,omitempty"`
	}

	result, err := mcpClient.ExecToolWithAny(
		"brave_web_search", 
		BraveInput{
			Query: "How to create a result type in golang",
			Count: 3,
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Tool execution result:", result)

}
