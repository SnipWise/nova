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

	type ToolInput struct {
		Name string `json:"name"`
	}
	result, err := mcpClient.ExecToolWithAny("hello_world_with_name", ToolInput{Name: "Alice"})

	if err != nil {
		panic(err)
	}

	fmt.Println("Tool execution result:", result)

}
