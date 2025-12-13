package mcptools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
)

// MCPClient wraps an MCP client connection with available tools
type MCPClient struct {
	mcpclient   *client.Client
	ToolsResult *mcp.ListToolsResult
	ctx         context.Context
}

func NewStdioMCPClient(ctx context.Context, command string, env []string, args ...string) (*MCPClient, error) {
	mcpClient, err := client.NewStdioMCPClient(
		command,
		env,
		args...,
	)
	if err != nil {
		return nil, err
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "nova",
		Version: "0.0.0",
	}
	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}
	toolsRequest := mcp.ListToolsRequest{}
	mcpTools, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, err
	}

	return &MCPClient{
		mcpclient:   mcpClient,
		ToolsResult: mcpTools,
		ctx:         ctx,
	}, nil

}

// NewStreamableHttpMCPClient creates and initializes a new MCP client over HTTP
func NewStreamableHttpMCPClient(ctx context.Context, mcpHostURL string) (*MCPClient, error) {
	mcpClient, err := client.NewStreamableHttpClient(
		mcpHostURL, // Use environment variable for MCP host
	)
	//defer mcpClient.Close()
	if err != nil {
		return nil, err
	}
	// Start the connection to the server
	err = mcpClient.Start(ctx)
	if err != nil {
		return nil, err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "nova",
		Version: "0.0.0",
	}
	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Streamable HTTP client connected & initialized with server!", result)
	//ui.Println(ui.Yellow, "Streamable HTTP client connected & initialized with server!")

	toolsRequest := mcp.ListToolsRequest{}
	mcpTools, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, err
	}

	return &MCPClient{
		mcpclient:   mcpClient,
		ToolsResult: mcpTools,
		ctx:         ctx,
	}, nil
}

// OpenAITools converts the MCP client's tools to OpenAI-compatible format
func (c *MCPClient) OpenAITools() []openai.ChatCompletionToolUnionParam {
	return ConvertMCPListToolsResultToOpenAITools(c.ToolsResult)
}

// OpenAIToolsWithFilter converts only the filtered MCP tools to OpenAI-compatible format
func (c *MCPClient) OpenAIToolsWithFilter(toolsFilter []string) []openai.ChatCompletionToolUnionParam {
	return ConvertMCPListToolsResultToOpenAIToolsWithFilter(c.ToolsResult, toolsFilter)
}

func (c *MCPClient) GetTools() []mcp.Tool {
	return c.ToolsResult.Tools
}

// IMPORTANT: TODO: TO BE TESTED
// GetToolsWithFilter returns only the tools that match the provided filter names
func (c *MCPClient) GetToolsWithFilter(toolsFilter []string) []mcp.Tool  {
	// Create a set for quick lookup of allowed tool names
	allowedTools := make(map[string]bool)
	for _, name := range toolsFilter {
		allowedTools[name] = true
	}

	// Filter tools
	filteredTools := &mcp.ListToolsResult{
		Tools: []mcp.Tool{},
	}
	for _, tool := range c.ToolsResult.Tools {
		if allowedTools[tool.Name] {
			filteredTools.Tools = append(filteredTools.Tools, tool)
		}
	}

	return filteredTools.Tools
}

// Close safely closes the MCP client connection
func (c *MCPClient) Close() error {
	if c.mcpclient != nil {
		return c.mcpclient.Close()
	}
	return nil
}

// CallTool executes a tool call with the given function name and JSON arguments
// func (c *MCPClient) CallTool(ctx context.Context, functionName string, arguments string) (*mcp.CallToolResult, error) {

// 	// Parse the tool arguments from JSON string
// 	var args map[string]any
// 	args, _ = conversion.JsonStringToMap(arguments)
// 	// TODO: check if this is useful for the request

// 	// NOTE: Call the MCP tool with the arguments
// 	request := mcp.CallToolRequest{}
// 	request.Params.Name = functionName
// 	request.Params.Arguments = args

// 	// NOTE: Call the tool using the MCP client
// 	toolResponse, err := c.mcpclient.CallTool(ctx, request)
// 	if err != nil {
// 		return nil, fmt.Errorf("error calling tool %s: %w", functionName, err)
// 	}
// 	if toolResponse == nil || len(toolResponse.Content) == 0 {
// 		return nil, fmt.Errorf("no content returned from tool %s", functionName)
// 	}

// 	return toolResponse, nil
// }

func (c *MCPClient) ExecToolWithMap(functionName string, input map[string]any) (string, error) {
	// NOTE: Call the MCP tool with the arguments
	request := mcp.CallToolRequest{}
	request.Params.Name = functionName
	request.Params.Arguments = input

	// NOTE: Call the tool using the MCP client
	toolResponse, err := c.mcpclient.CallTool(c.ctx, request)
	if err != nil {
		return "", fmt.Errorf("error calling tool %s: %w", functionName, err)
	}
	if toolResponse == nil || len(toolResponse.Content) == 0 {
		return "", fmt.Errorf("no content returned from tool %s", functionName)
	}

	// Take the first content item and return its text
	resultContent := toolResponse.Content[0].(mcp.TextContent).Text
	return resultContent, nil
}

func (c *MCPClient) ExecToolWithString(functionName string, input string) (string, error) {
	// Parse the tool arguments from JSON string
	var args map[string]any
	args, err := conversion.JsonStringToMap(input)
	if err != nil {
		return "", fmt.Errorf("error converting input string to map: %w", err)
	}
	return c.ExecToolWithMap(functionName, args)
}

func (c *MCPClient) ExecToolWithAny(functionName string, input any) (string, error) {
	// Convert input to map[string]any
	args, err := conversion.AnyToMap(input)
	if err != nil {
		return "", fmt.Errorf("error converting input to map: %w", err)
	}
	return c.ExecToolWithMap(functionName, args)
}

func Exec[I, O any](mcpClient *MCPClient, functionName string, input I) (O, error) {
	var output O
	// Convert input to map[string]any
	args, err := conversion.AnyToMap(input)
	if err != nil {
		return output, fmt.Errorf("error converting input to map: %w", err)
	}
	resultContent, err := mcpClient.ExecToolWithMap(functionName, args)
	if err != nil {
		return output, err
	}

	// Convert resultContent (string) to output type O
	output, err = conversion.FromJSON[O](resultContent)
	if err != nil {
		return output, fmt.Errorf("error converting output from JSON string: %w", err)
	}

	return output, nil
}


func ConvertMCPToolsToOpenAITools(tools []mcp.Tool) []openai.ChatCompletionToolUnionParam {
	openAITools := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {

		openAITools[i] = openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        tool.Name,
			Description: openai.String(tool.Description),
			Parameters: shared.FunctionParameters{
				"type":       "object",
				"properties": tool.InputSchema.Properties,
				"required":   tool.InputSchema.Required,
			},
		},
		)
	}
	return openAITools
}

func ConvertMCPToolsToOpenAIToolsWithFilter(tools []mcp.Tool, toolsFilter []string) []openai.ChatCompletionToolUnionParam {
	// Create a set for quick lookup of allowed tool names
	allowedTools := make(map[string]bool)
	for _, name := range toolsFilter {
		allowedTools[name] = true
	}

	// Filter tools and convert to OpenAI format
	var openAITools []openai.ChatCompletionToolUnionParam
	for _, tool := range tools {
		if allowedTools[tool.Name] {
			openAITools = append(openAITools, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters: shared.FunctionParameters{
					"type":       "object",
					"properties": tool.InputSchema.Properties,
					"required":   tool.InputSchema.Required,
				},
			},
			))
		}
	}
	return openAITools
}



// ConvertMCPListToolsResultToOpenAITools transforms MCP tool definitions into OpenAI tool format
func ConvertMCPListToolsResultToOpenAITools(tools *mcp.ListToolsResult) []openai.ChatCompletionToolUnionParam {
	openAITools := make([]openai.ChatCompletionToolUnionParam, len(tools.Tools))
	for i, tool := range tools.Tools {

		openAITools[i] = openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        tool.Name,
			Description: openai.String(tool.Description),
			Parameters: shared.FunctionParameters{
				"type":       "object",
				"properties": tool.InputSchema.Properties,
				"required":   tool.InputSchema.Required,
			},
		},
		)
	}
	return openAITools
}

// ConvertMCPListToolsResultToOpenAIToolsWithFilter transforms filtered MCP tool definitions into OpenAI tool format
func ConvertMCPListToolsResultToOpenAIToolsWithFilter(tools *mcp.ListToolsResult, toolsFilter []string) []openai.ChatCompletionToolUnionParam {
	// Create a set for quick lookup of allowed tool names
	allowedTools := make(map[string]bool)
	for _, name := range toolsFilter {
		allowedTools[name] = true
	}

	// Filter tools and convert to OpenAI format
	var openAITools []openai.ChatCompletionToolUnionParam
	for _, tool := range tools.Tools {
		if allowedTools[tool.Name] {
			openAITools = append(openAITools, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters: shared.FunctionParameters{
					"type":       "object",
					"properties": tool.InputSchema.Properties,
					"required":   tool.InputSchema.Required,
				},
			},
			))
		}
	}
	return openAITools
}
