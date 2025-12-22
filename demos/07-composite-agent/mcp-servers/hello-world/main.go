package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)


func main() {

	// Create MCP server
	s := server.NewMCPServer(
		"mcp-hello-world",
		"0.0.0",
	)
	
	helloWorldTool := mcp.NewTool("hello_world",
		mcp.WithDescription("A simple hello world tool"),
	)
	s.AddTool(helloWorldTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Hello, World! This is MCP server responding."), nil
	})

	helloWorldToolWithName := mcp.NewTool("hello_world_with_name",
		mcp.WithDescription("A simple hello world tool that greets a user by name"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the user to greet"),
		),
	)
	s.AddTool(helloWorldToolWithName, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		nameArg, exists := args["name"]
		if !exists || nameArg == nil {
			return nil, fmt.Errorf("missing required parameter 'name'")
		}

		name, ok := nameArg.(string)
		if !ok {
			return nil, fmt.Errorf("parameter 'name' must be a string")
		}

		greeting := fmt.Sprintf("Hello, %s! This is MCP server responding.", name)
		return mcp.NewToolResultText(greeting), nil
	})

	// Start the HTTP server
	httpPort := os.Getenv("MCP_HTTP_PORT")
	if httpPort == "" {
		httpPort = "9090"
	}

	log.Println("MCP 👋 Hello World 🌍 Server is running on port", httpPort)

	// Create a custom mux to handle both MCP and health endpoints
	mux := http.NewServeMux()

	// Add healthcheck endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// Add MCP endpoint
	httpServer := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp"),
	)

	// Register MCP handler with the mux
	mux.Handle("/mcp", httpServer)

	// Start the HTTP server with custom mux
	log.Fatal(http.ListenAndServe(":"+httpPort, mux))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status": "healthy",
		"server": "mcp-hello-world-server",
	}
	json.NewEncoder(w).Encode(response)
}
