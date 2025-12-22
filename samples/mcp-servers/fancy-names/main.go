package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var projectAdjectives = []string{
	"Cosmic", "Stellar", "Quantum", "Digital", "Cyber", "Neural", "Swift", "Dynamic",
	"Epic", "Infinite", "Mystic", "Turbo", "Ultra", "Mega", "Hyper", "Super",
	"Smart", "Rapid", "Fusion", "Phoenix", "Thunder", "Crystal", "Golden", "Silver",
}

var projectNouns = []string{
	"Phoenix", "Nebula", "Horizon", "Genesis", "Matrix", "Vertex", "Nexus", "Prism",
	"Pulse", "Echo", "Atlas", "Titan", "Omega", "Alpha", "Delta", "Sigma",
	"Core", "Hub", "Link", "Wave", "Flow", "Spark", "Blaze", "Storm",
}

var variablePrefixes = []string{
	"current", "active", "selected", "main", "primary", "new", "temp", "cached",
	"stored", "parsed", "formatted", "validated", "processed", "filtered", "sorted",
	"updated", "initialized", "configured", "loaded", "ready",
}

var variableTypes = []string{
	"user", "data", "config", "item", "list", "map", "set", "queue", "stack",
	"buffer", "stream", "handler", "manager", "service", "controller", "model",
	"view", "context", "session", "request", "response", "payload", "result",
}

var functionVerbs = []string{
	"create", "build", "generate", "fetch", "load", "save", "update", "delete",
	"process", "validate", "parse", "format", "transform", "convert", "filter",
	"sort", "merge", "split", "join", "extract", "compute", "calculate", "handle",
}

var functionObjects = []string{
	"User", "Data", "Config", "Item", "List", "Record", "Entry", "Document",
	"Message", "Event", "Task", "Job", "Request", "Response", "Payload", "Result",
	"Session", "Context", "Connection", "Transaction", "Query", "Report", "Summary",
}

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"mcp-fancy-names",
		"1.0.0",
	)
	

	// Tool 1: Generate project names
	projectNameTool := mcp.NewTool("generate_project_name",
		mcp.WithDescription("Generate fancy project names"),
		mcp.WithNumber("count",
			mcp.Description("Number of project names to generate (default: 1)"),
		),
		mcp.WithString("style",
			mcp.Description("Style of project name: 'compound' (adjective+noun), 'single' (noun only), or 'any' (default: 'compound')"),
		),
	)
	s.AddTool(projectNameTool, handleGenerateProjectName)

	// Tool 2: Generate variable names
	variableNameTool := mcp.NewTool("generate_variable_name",
		mcp.WithDescription("Generate fancy variable names following common naming conventions"),
		mcp.WithNumber("count",
			mcp.Description("Number of variable names to generate (default: 1)"),
		),
		mcp.WithString("case_style",
			mcp.Description("Naming convention: 'camelCase', 'snake_case', 'PascalCase', or 'any' (default: 'camelCase')"),
		),
	)
	s.AddTool(variableNameTool, handleGenerateVariableName)

	// Tool 3: Generate function names
	functionNameTool := mcp.NewTool("generate_function_name",
		mcp.WithDescription("Generate fancy function names following common naming conventions"),
		mcp.WithNumber("count",
			mcp.Description("Number of function names to generate (default: 1)"),
		),
		mcp.WithString("case_style",
			mcp.Description("Naming convention: 'camelCase', 'snake_case', 'PascalCase', or 'any' (default: 'camelCase')"),
		),
	)
	s.AddTool(functionNameTool, handleGenerateFunctionName)

	// Start the HTTP server
	httpPort := os.Getenv("MCP_HTTP_PORT")
	if httpPort == "" {
		httpPort = "9091"
	}

	log.Println("MCP ✨ Fancy Names 🎨 Server is running on port", httpPort)

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

func handleGenerateProjectName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	count := 1
	if countArg, exists := args["count"]; exists && countArg != nil {
		if c, ok := countArg.(float64); ok {
			count = int(c)
		}
	}

	style := "compound"
	if styleArg, exists := args["style"]; exists && styleArg != nil {
		if s, ok := styleArg.(string); ok {
			style = s
		}
	}

	var names []string
	for i := 0; i < count; i++ {
		var name string
		if style == "single" {
			name = projectNouns[rand.Intn(len(projectNouns))]
		} else if style == "any" {
			if rand.Intn(2) == 0 {
				name = projectNouns[rand.Intn(len(projectNouns))]
			} else {
				name = projectAdjectives[rand.Intn(len(projectAdjectives))] + projectNouns[rand.Intn(len(projectNouns))]
			}
		} else { // compound
			name = projectAdjectives[rand.Intn(len(projectAdjectives))] + projectNouns[rand.Intn(len(projectNouns))]
		}
		names = append(names, name)
	}

	result := strings.Join(names, "\n")
	return mcp.NewToolResultText(result), nil
}

func handleGenerateVariableName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	count := 1
	if countArg, exists := args["count"]; exists && countArg != nil {
		if c, ok := countArg.(float64); ok {
			count = int(c)
		}
	}

	caseStyle := "camelCase"
	if caseArg, exists := args["case_style"]; exists && caseArg != nil {
		if s, ok := caseArg.(string); ok {
			caseStyle = s
		}
	}

	var names []string
	for i := 0; i < count; i++ {
		prefix := variablePrefixes[rand.Intn(len(variablePrefixes))]
		varType := variableTypes[rand.Intn(len(variableTypes))]

		var name string
		actualCase := caseStyle
		if caseStyle == "any" {
			cases := []string{"camelCase", "snake_case", "PascalCase"}
			actualCase = cases[rand.Intn(len(cases))]
		}

		switch actualCase {
		case "snake_case":
			name = prefix + "_" + varType
		case "PascalCase":
			name = strings.Title(prefix) + strings.Title(varType)
		default: // camelCase
			name = prefix + strings.Title(varType)
		}
		names = append(names, name)
	}

	result := strings.Join(names, "\n")
	return mcp.NewToolResultText(result), nil
}

func handleGenerateFunctionName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	count := 1
	if countArg, exists := args["count"]; exists && countArg != nil {
		if c, ok := countArg.(float64); ok {
			count = int(c)
		}
	}

	caseStyle := "camelCase"
	if caseArg, exists := args["case_style"]; exists && caseArg != nil {
		if s, ok := caseArg.(string); ok {
			caseStyle = s
		}
	}

	var names []string
	for i := 0; i < count; i++ {
		verb := functionVerbs[rand.Intn(len(functionVerbs))]
		object := functionObjects[rand.Intn(len(functionObjects))]

		var name string
		actualCase := caseStyle
		if caseStyle == "any" {
			cases := []string{"camelCase", "snake_case", "PascalCase"}
			actualCase = cases[rand.Intn(len(cases))]
		}

		switch actualCase {
		case "snake_case":
			name = verb + "_" + strings.ToLower(object)
		case "PascalCase":
			name = strings.Title(verb) + object
		default: // camelCase
			name = verb + object
		}
		names = append(names, name)
	}

	result := strings.Join(names, "\n")
	return mcp.NewToolResultText(result), nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status": "healthy",
		"server": "mcp-fancy-names-server",
	}
	json.NewEncoder(w).Encode(response)
}
