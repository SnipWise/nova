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

	type DuckDuckGoSearchInput struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results,omitempty"`
	}

	type DuckDuckGoFetchInput struct {
		Url string `json:"url"`
	}

	result, err := mcpClient.ExecToolWithAny(
		"search",
		DuckDuckGoSearchInput{
			Query:      "How to create a result type in golang",
			MaxResults: 3,
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("📝 Tool execution result:", result)

	result, err = mcpClient.ExecToolWithAny(
		"fetch_content",
		DuckDuckGoFetchInput{
			Url: "https://pkg.go.dev/github.com/alecthomas/types/result",
		},
	)

	if err != nil {
		panic(err)
	}
	fmt.Println("🌍 Tool execution result:", result)

}
