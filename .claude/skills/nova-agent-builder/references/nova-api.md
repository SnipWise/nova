# Nova SDK API Reference

Complete API reference for Nova SDK - a Go framework for building AI agents.

## Core Imports

```go
import (
    // Core Agent Types
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/base"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/agents/crew"
    "github.com/snipwise/nova/nova-sdk/agents/server"
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
    "github.com/snipwise/nova/nova-sdk/agents/remote"

    // Messages & Models
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"

    // UI & Utilities
    "github.com/snipwise/nova/nova-sdk/ui/display"
    "github.com/snipwise/nova/nova-sdk/ui/prompt"
    "github.com/snipwise/nova/nova-sdk/ui/spinner"
    "github.com/snipwise/nova/nova-sdk/toolbox/env"
    "github.com/snipwise/nova/nova-sdk/toolbox/logger"
    "github.com/snipwise/nova/nova-sdk/toolbox/conversion"

    // MCP Integration
    "github.com/snipwise/nova/nova-sdk/mcptools"
)
```

---

## Configuration

### agents.Config

Core configuration for all agent types:

```go
type Config struct {
    Name                    string  // Agent identifier (optional)
    Description             string  // Agent purpose (optional)
    SystemInstructions      string  // Behavior/role definition
    EngineURL              string  // Model inference engine URL (required)
    APIKey                 string  // Authentication key (optional)
    KeepConversationHistory bool   // Conversation memory management (default: true)
}
```

**KeepConversationHistory:**
- `true`: Messages accumulate across calls (stateful, contextual)
- `false`: Only system message persists (stateless, context resets)

### models.Config

Model configuration with fluent builder:

```go
type Config struct {
    Name              string    // Model name (required)
    Temperature       *float64  // Sampling temperature (0.0-1.0)
    TopP              *float64  // Nucleus sampling
    TopK              *int64    // Top-K sampling
    MinP              *float64  // Minimum probability
    MaxTokens         *int64    // Max completion tokens
    FrequencyPenalty  *float64  // Penalize repeated tokens
    PresencePenalty   *float64  // Penalize token presence
    RepeatPenalty     *float64  // Penalize repetitions
    Seed              *int64    // Deterministic sampling seed
    Stop              []string  // Stop sequences
    N                 *int64    // Number of completions
    ParallelToolCalls *bool     // Enable parallel tool execution
    ReasoningEffort   *string   // Reasoning depth (for reasoning models)
}

// Pointer Helpers
models.Float64(0.7)  // Returns *float64
models.Int(2000)     // Returns *int
models.Bool(true)    // Returns *bool

// Fluent Builder (recommended)
config := models.NewConfig("ai/qwen2.5:1.5B-F16").
    WithTemperature(0.7).
    WithMaxTokens(2000).
    WithToolChoiceAuto().
    WithParallelToolCalls(true)

// Reasoning Effort Constants
models.ReasoningEffortNone
models.ReasoningEffortMinimal
models.ReasoningEffortLow
models.ReasoningEffortMedium   // Default
models.ReasoningEffortHigh
models.ReasoningEffortXHigh

// Predefined Config Templates
models.DeterministicConfig("model-id")  // Temperature 0.0
models.CreativeConfig("model-id")       // Temperature 0.9
models.BalancedConfig("model-id")       // Temperature 0.5
```

---

## Chat Agent

Simple conversational agent with automatic history management.

### Creation

```go
agent, err := chat.NewAgent(
    ctx,
    agents.Config{
        Name:                    "helpful-assistant",
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "You are a helpful assistant.",
        KeepConversationHistory: true,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
        MaxTokens:   models.Int(2000),
    },
)
```

### Methods

```go
// Non-Streaming Completion
type CompletionResult struct {
    Response     string
    FinishReason string
}

result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Hello"},
})

// Streaming Completion
type StreamCallback func(chunk string, finishReason string) error

result, err := agent.GenerateStreamCompletion(
    []messages.Message{{Role: roles.User, Content: "Hello"}},
    func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil // Return error to stop streaming
    },
)

// With Reasoning (for reasoning-capable models)
type ReasoningResult struct {
    Response     string
    Reasoning    string
    FinishReason string
}

result, err := agent.GenerateCompletionWithReasoning(messages)

// Streaming with Reasoning
result, err := agent.GenerateStreamCompletionWithReasoning(
    messages,
    func(reasoningChunk, finishReason string) error {
        fmt.Print(reasoningChunk)
        return nil
    },
    func(responseChunk, finishReason string) error {
        fmt.Print(responseChunk)
        return nil
    },
)

// History Management
messages := agent.GetMessages()          // Get conversation history
agent.ResetMessages()                    // Clear history
contextSize := agent.GetContextSize()    // Get token count
json, _ := agent.ExportMessagesToJSON()  // Export to JSON
```

---

## RAG Agent

Retrieval-Augmented Generation with vector storage.

### Creation

```go
agent, err := rag.NewAgent(
    ctx,
    agents.Config{
        Name:      "rag-assistant",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    },
    models.Config{
        Name: "ai/mxbai-embed-large", // Default embedding model
    },
)
```

### Methods

```go
// Embedding Generation
embedding, err := agent.GenerateEmbedding(content)
dimension := agent.GetEmbeddingDimension()

// Document Indexing
err = agent.SaveEmbedding("Document content chunk")
err = agent.SaveEmbeddingIntoMemoryVectorStore("Chunk with metadata")

// Persistence
err = agent.LoadStore("./store/data.json")
err = agent.PersistStore("./store/data.json")
exists := agent.StoreFileExists("./store/data.json")

// Similarity Search
type VectorRecord struct {
    ID         string
    Prompt     string
    Embedding  []float64
    Metadata   map[string]any
    Similarity float64  // Cosine similarity score
}

records, err := agent.SearchSimilar(query, 0.6)           // threshold: 0.0-1.0
records, err := agent.SearchTopN(query, 0.6, 5)           // top 5 results

// Common Pattern (with persistence)
if agent.StoreFileExists("./store/data.json") {
    agent.LoadStore("./store/data.json")
} else {
    // Index documents
    for _, doc := range documents {
        agent.SaveEmbedding(doc)
    }
    agent.PersistStore("./store/data.json")
}
```

### Chunking Utilities

```go
import "github.com/snipwise/nova/nova-sdk/agents/rag/chunks"

// Split by markdown sections
pieces := chunks.SplitMarkdownBySections(markdownText)

// Chunk by size
chunked := chunks.ChunkText(text, 512, 64) // size: 512, overlap: 64
```

---

## Tools Agent

Function calling with sequential or parallel execution.

### Tool Definition

```go
tool := tools.NewTool("calculate_sum").
    SetDescription("Calculate the sum of two numbers").
    AddParameter("a", "number", "First number", true).        // required
    AddParameter("b", "number", "Second number", true).
    AddParameter("precision", "integer", "Decimal places", false) // optional

// Enum Parameter
tool.AddEnumParameter("status", "string", "Status", []string{"active", "inactive"}, true)
```

### Agent Creation

```go
agent, err := tools.NewAgent(
    ctx,
    agents.Config{
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "You are an AI with access to tools.",
    },
    models.Config{
        Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature:       models.Float64(0.0), // ALWAYS 0.0 for tools
        ParallelToolCalls: models.Bool(true),
    },
    tools.WithTools([]*tools.Tool{tool1, tool2}),
    // OR
    tools.WithOpenAITools(openaiTools),
    // OR
    tools.WithMCPTools(mcpTools),
)
```

### Methods

```go
// Execute Function Callback
type ToolCallback func(functionName string, arguments string) (string, error)

executeFunction := func(functionName string, arguments string) (string, error) {
    switch functionName {
    case "calculate_sum":
        var args struct {
            A float64 `json:"a"`
            B float64 `json:"b"`
        }
        json.Unmarshal([]byte(arguments), &args)
        return fmt.Sprintf(`{"result": %g}`, args.A + args.B), nil
    default:
        return `{"error": "Unknown function"}`, fmt.Errorf("unknown: %s", functionName)
    }
}

// Confirmation Callback (Human-in-the-Loop)
type ConfirmationCallback func(functionName string, arguments string) ConfirmationResponse

const (
    Confirmed ConfirmationResponse = iota
    Denied
    Quit
)

// Sequential Execution (loop until no more tool calls)
type ToolCallResult struct {
    FinishReason         string
    Results              []string
    LastAssistantMessage string
}

result, err := agent.DetectToolCallsLoop(messages, executeFunction)

// With Confirmation
result, err := agent.DetectToolCallsLoopWithConfirmation(
    messages,
    executeFunction,
    func(functionName, arguments string) tools.ConfirmationResponse {
        fmt.Printf("Execute %s(%s)? [y/n/q]: ", functionName, arguments)
        // ... prompt user ...
        return tools.Confirmed
    },
)

// Parallel Execution (single round, multiple tools)
result, err := agent.DetectParallelToolCalls(messages, executeFunction)

// With Confirmation
result, err := agent.DetectParallelToolCallsWithConfirmation(
    messages,
    executeFunction,
    confirmationCallback,
)

// Streaming Variants
result, err := agent.DetectToolCallsLoopStream(
    messages,
    executeFunction,
    func(chunk, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)

// State Management
state := agent.GetLastStateToolCalls()
agent.ResetLastStateToolCalls()
```

---

## Structured Agent

Type-safe structured output using Go generics.

### Creation

```go
type Country struct {
    Name       string   `json:"name"`
    Capital    string   `json:"capital"`
    Population int      `json:"population"`
    Languages  []string `json:"languages"`
}

agent, err := structured.NewAgent[Country](
    ctx,
    agents.Config{
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "Answer questions about countries.",
    },
    models.Config{
        Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature: models.Float64(0.0), // Recommended for structured output
    },
)
```

### Methods

```go
// Generate Structured Data
response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Tell me about Canada."},
})

// response is *Country with guaranteed structure
fmt.Println(response.Capital) // "Ottawa"
```

---

## Compressor Agent

Context compression for managing conversation history.

### Creation

```go
agent, err := compressor.NewAgent(
    ctx,
    agents.Config{
        Name:      "compressor",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        // RECOMMENDED: Use built-in instructions
        SystemInstructions: compressor.Instructions.Effective,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0), // CRITICAL: Always 0.0 for consistency
    },
    // RECOMMENDED: Use built-in prompts
    compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
)
```

### Predefined Configurations

```go
// Instructions (SystemInstructions)
compressor.Instructions.Expert      // Most sophisticated compression
compressor.Instructions.Effective   // Balanced (RECOMMENDED)
compressor.Instructions.Basic       // Simple compression

// Prompts (WithCompressionPrompt)
compressor.Prompts.UltraShort   // Maximum token reduction (RECOMMENDED)
compressor.Prompts.Minimalist   // Very concise
compressor.Prompts.Balanced     // Balance detail/brevity
compressor.Prompts.Detailed     // Preserve more information
```

### Methods

```go
type CompressionResult struct {
    CompressedText string
    FinishReason   string
}

type StreamCallback func(chunk string, finishReason string) error

// Compress Text
result, err := agent.CompressContext(messagesList)

// Streaming Compression
result, err := agent.CompressContextStream(
    messagesList,
    func(chunk, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)

// Set Custom Prompt
agent.SetCompressionPrompt("Custom compression instructions...")

// Common Pattern (context management)
if chatAgent.GetContextSize() > threshold {
    compressed, _ := compressorAgent.CompressContextStream(
        chatAgent.GetMessages(),
        streamCallback,
    )
    chatAgent.ResetMessages()
    chatAgent.AddMessage(roles.System, compressed.CompressedText)
}
```

---

## Orchestrator Agent

Topic/intent detection for routing in multi-agent systems.

### Creation

```go
agent, err := orchestrator.NewAgent(
    ctx,
    agents.Config{
        Name:      "orchestrator",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: `
            Identify the main topic of discussion.
            Possible topics: Technology, Health, Science, Business.
            Respond in JSON with field 'topic_discussion'.
        `,
    },
    models.Config{
        Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
        Temperature: models.Float64(0.0), // Deterministic for consistent routing
    },
)
```

### Methods

```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}

// Identify Intent
intent, finishReason, err := agent.IdentifyIntent(messages)
fmt.Println(intent.TopicDiscussion) // "Technology"

// Identify from Text
topic, err := agent.IdentifyTopicFromText("I want to learn about Python programming")
fmt.Println(topic) // "Technology"
```

---

## Crew Agent

Multi-agent collaboration with routing and shared capabilities.

### Creation

```go
// Create agent crew
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "analyst": analystAgent,
    "expert":  expertAgent,
}

// Routing function
matchAgentFunction := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding", "programming":
        return "coder"
    case "analysis", "data":
        return "analyst"
    default:
        return "expert"
    }
}

// Crew Agent
agent, err := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "expert"),             // Default: expert
    crew.WithMatchAgentIdToTopicFn(matchAgentFunction),  // Routing
    crew.WithToolsAgent(toolsAgent),                     // Optional: Add tools
    crew.WithRagAgent(ragAgent),                         // Optional: Add RAG
    crew.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 7), // With config
    crew.WithCompressorAgent(compressorAgent),           // Optional: Add compression
    crew.WithCompressorAgentAndContextSize(compressorAgent, 8500), // With threshold
    crew.WithOrchestratorAgent(orchestratorAgent),       // Optional: Auto-routing
    crew.WithExecuteFn(executeFunction),                 // Optional: Tool execution
    crew.WithConfirmationPromptFn(confirmationCallback), // Optional: Confirmation
)

// Single Agent Mode (crew with one agent)
agent, err := crew.NewAgent(ctx, crew.WithSingleAgent(chatAgent))
```

### Methods

```go
// Crew Management
agents := agent.GetChatAgents()
agent.SetChatAgents(chatAgents)
agent.AddChatAgentToCrew("new-id", chatAgent)
agent.RemoveChatAgentFromCrew("agent-id")

// Agent Selection
currentId := agent.GetSelectedAgentId()
agent.SetSelectedAgentId("analyst")

// All completion methods delegate to current chat agent
result, err := agent.GenerateCompletion(messages)
result, err := agent.GenerateStreamCompletion(messages, callback)
```

---

## Server Agent

HTTP/REST API agent with SSE streaming.

### Creation

```go
agent, err := server.NewAgent(
    ctx,
    agents.Config{
        Name:               "server-agent",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "System prompt",
    },
    models.Config{
        Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature: models.Float64(0.4),
    },
    ":3500",         // Port
    executeFunction, // Optional: for tools
)

// Attach optional capabilities
agent.SetToolsAgent(toolsAgent)
agent.SetRagAgent(ragAgent)
agent.SetCompressorAgent(compressorAgent)
agent.SetConfirmationPromptFn(confirmationCallback)
```

### HTTP Endpoints

```
POST   /completion                     - SSE streaming completion
POST   /completion/stop                - Stop streaming
POST   /memory/reset                   - Clear conversation history
GET    /memory/messages/list           - Get all messages
GET    /memory/messages/context-size   - Get context size
POST   /operation/validate             - Validate tool call
POST   /operation/cancel               - Cancel tool call
POST   /operation/reset                - Reset all operations
GET    /models                         - Get model information
GET    /health                         - Health check
```

### Methods

```go
// Start Server (blocking)
log.Fatal(agent.StartServer())

// Configuration
agent.SetPort(":8080")
port := agent.GetPort()

// Dual-Mode (CLI + Server)
// Use StreamCompletion for CLI mode
result, err := agent.StreamCompletion(question, func(chunk, finishReason string) error {
    fmt.Print(chunk)
    return nil
})
```

---

## Crew Server Agent

HTTP server for multi-agent crew.

### Creation

```go
agent, err := crewserver.NewAgent(
    ctx,
    // All crew.NewAgent options, plus:
    crewserver.WithPort(3500),
    crewserver.WithAgentCrew(agentCrew, "expert"),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithRagAgent(ragAgent),
    crewserver.WithCompressorAgent(compressorAgent),
)

// Start Server
log.Fatal(agent.StartServer())
```

**Endpoints:** Same as Server Agent

---

## Remote Agent

HTTP client for Server/Crew Server agents.

### Creation

```go
agent, err := remote.NewAgent(
    ctx,
    agents.Config{
        Name: "remote-client",
    },
    "http://localhost:3500", // Server base URL
)
```

### Methods

```go
// Completion (same as Chat Agent)
result, err := agent.GenerateCompletion(messages)
result, err := agent.GenerateStreamCompletion(messages, callback)

// Remote Operations
err = agent.ValidateOperation("operation-id")
err = agent.CancelOperation("operation-id")
err = agent.ResetOperations()

// Tool Call Callback
type ToolCallCallback func(operationID string, message string) error

agent.SetToolCallCallback(func(operationID, message string) error {
    fmt.Printf("Tool call: %s - %s\n", operationID, message)
    // Validate or cancel via HTTP endpoints
    return nil
})

// Server Information
info, err := agent.GetModelsInfo()
health, err := agent.GetHealth()
isHealthy := agent.IsHealthy()

// History Management (local only, not synced with server)
agent.GetMessages()
agent.ResetMessages()
agent.GetContextSize()
```

---

## Messages

### Roles

```go
roles.System    // System message
roles.User      // User message
roles.Assistant // Assistant message
roles.Developer // Developer message (for debugging)
roles.Tool      // Tool message (function responses)
```

### Message Structure

```go
type Message struct {
    Role    roles.Role
    Content string
}

// Example
message := messages.Message{
    Role:    roles.User,
    Content: "What is the capital of France?",
}
```

### Conversion Helpers

```go
// Convert to OpenAI SDK types (internal use)
openaiMsg := messages.ConvertToOpenAIMessage(msg)
openaiMsgs := messages.ConvertToOpenAIMessages(msgs)

// Convert from OpenAI SDK types
novaMsg := messages.ConvertFromOpenAIMessage(openaiMsg)
novaMsgs := messages.ConvertFromOpenAIMessages(openaiMsgs)
```

---

## UI Utilities

### Display

```go
import "github.com/snipwise/nova/nova-sdk/ui/display"

// Basic Output
display.Print("Hello")
display.Println("Hello")
display.Printf("Count: %d", 42)

// Colors
display.Success("Operation completed")
display.Error("An error occurred")
display.Warning("Be careful")
display.Info("FYI")

// Styled Output
display.Header("Section Header")
display.Title("Main Title")
display.Box("Boxed message")
display.Banner("Application Name")

// Markdown Rendering
display.Markdown("# Title\n\nParagraph with **bold**")
display.MarkdownChunk(chunk) // For streaming

// Color Constants
display.ColorRed, display.ColorGreen, display.ColorYellow, display.ColorBlue
display.ColorBrightRed, display.ColorBrightGreen, display.ColorBrightCyan
```

### Prompts

```go
import "github.com/snipwise/nova/nova-sdk/ui/prompt"

// Input Prompt
input := prompt.New("Enter your name").
    SetDefault("User").
    SetValidator(func(s string) error {
        if len(s) < 3 {
            return errors.New("name too short")
        }
        return nil
    })
result, err := input.Run()

// Confirmation Prompt
confirm := prompt.NewConfirm("Continue?").SetDefault(true)
result, err := confirm.Run() // true/false

// Select Prompt
select := prompt.NewSelect("Choose option", []prompt.Choice{
    {Label: "Option 1", Value: "opt1"},
    {Label: "Option 2", Value: "opt2"},
})
result, err := select.Run()

// Multi-Choice Prompt
multi := prompt.NewMultiChoice("Select multiple", choices).
    SetDefaults([]string{"opt1", "opt2"})
results, err := multi.Run() // []string

// Tool Confirmation (built-in helper)
response := prompt.AskToolConfirmation(functionName, arguments)
// Returns: tools.Confirmed, tools.Denied, or tools.Quit

// Edit Prompts
text, err := prompt.RunWithEdit("Edit text", initialContent)
multiline, err := prompt.RunWithMultilineEdit("Edit", content)
```

### Spinner

```go
import "github.com/snipwise/nova/nova-sdk/ui/spinner"

// Create Spinner
sp := spinner.New("Loading").
    SetFrames(spinner.FramesBraille).
    SetDelay(100 * time.Millisecond)

// Start/Stop
sp.Start()
time.Sleep(2 * time.Second)
sp.Success("Done!")
// OR
sp.Fail("Error!")
// OR
sp.Stop()

// Dynamic Updates
sp.UpdatePrefix("Processing")
sp.UpdateSuffix("50%")

// State
state := sp.State()      // StateIdle, StateRunning, StateStopped
running := sp.IsRunning()

// Predefined Frames
spinner.FramesBraille
spinner.FramesDots
spinner.FramesASCII
spinner.FramesProgressive
spinner.FramesArrows
spinner.FramesCircle
spinner.FramesPulsingStar
```

---

## Toolbox Utilities

### Environment Variables

```go
import "github.com/snipwise/nova/nova-sdk/toolbox/env"

// Get with default fallback
engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
modelID := env.GetEnvOrDefault("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")

// Type-specific helpers
port := env.GetEnvIntOrDefault("HTTP_PORT", 3500)
enabled := env.GetEnvBoolOrDefault("FEATURE_ENABLED", false)
```

### Logger

```go
import "github.com/snipwise/nova/nova-sdk/toolbox/logger"

// Create Logger
log := logger.GetLoggerFromEnv()
log := logger.GetLoggerFromEnvWithPrefix("MyAgent")

// Log Levels
type LogLevel int

const (
    LevelDebug LogLevel = iota
    LevelInfo
    LevelWarn
    LevelError
    LevelNone
)

// Usage
log.SetLevel(logger.LevelInfo)
log.Debug("Debug message", "key", "value")
log.Info("Info message")
log.Warn("Warning")
log.Error("Error occurred")

// Environment Variable
// NOVA_LOG_LEVEL=debug|info|warn|error|none
```

### Conversion

```go
import "github.com/snipwise/nova/nova-sdk/toolbox/conversion"

// JSON Helpers
pretty, err := conversion.PrettyPrint(jsonString)
data, err := conversion.JsonStringToMap(jsonString)
data, err := conversion.AnyToMap(anyValue)

// Generic JSON Parsing
type MyStruct struct {
    Name string `json:"name"`
}
result, err := conversion.FromJSON[MyStruct](jsonString)
```

---

## MCP Integration

Model Context Protocol (MCP) integration for external tools.

### MCP Client

```go
import "github.com/snipwise/nova/nova-sdk/mcptools"

// STDIO Transport
mcpClient, err := mcptools.NewStdioMCPClient(
    ctx,
    "uvx",                          // command
    []string{"PATH=/usr/bin"},      // env
    "mcp-server-fetch",             // args...
)

// HTTP Transport
mcpClient, err := mcptools.NewStreamableHttpMCPClient(
    ctx,
    "http://localhost:8000/mcp",
)
```

### Tool Conversion

```go
// Get Tools
tools := mcpClient.GetTools()
filtered := mcpClient.GetToolsWithFilter([]string{"search", "fetch"})

// Convert to OpenAI Format
openaiTools := mcpClient.OpenAITools()
openaiFiltered := mcpClient.OpenAIToolsWithFilter([]string{"search"})

// Use with Tools Agent
toolsAgent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithMCPTools(mcpClient.GetTools()),
)
```

### Tool Execution

```go
// Execute with Map
result, err := mcpClient.ExecToolWithMap("search", map[string]any{
    "query": "golang tutorials",
})

// Execute with JSON String
result, err := mcpClient.ExecToolWithString("search", `{"query":"golang"}`)

// Execute with Any
result, err := mcpClient.ExecToolWithAny("search", searchParams)

// Generic Execution (type-safe)
type SearchInput struct {
    Query string `json:"query"`
}
type SearchOutput struct {
    Results []string `json:"results"`
}

output, err := mcptools.Exec[SearchInput, SearchOutput](
    mcpClient,
    "search",
    SearchInput{Query: "golang"},
)
```

---

## Recommended Models

| Use Case | Model | Temperature | Notes |
|----------|-------|-------------|-------|
| **Chat** | `ai/qwen2.5:1.5B-F16` | 0.0-0.8 | Good quality/speed ratio |
| **Embeddings** | `ai/mxbai-embed-large` | N/A | Default for RAG |
| **Tools** | `hf.co/menlo/jan-nano-gguf:q4_k_m` | **0.0** | Function calling support |
| **Structured** | `hf.co/menlo/jan-nano-gguf:q4_k_m` | **0.0** | Deterministic JSON |
| **Compressor** | `ai/qwen2.5:1.5B-F16` | **0.0** | Consistent compression |
| **Orchestrator** | `hf.co/menlo/lucy-gguf:q4_k_m` | **0.0** | Routing decisions |
| **Reasoning** | `hf.co/menlo/lucy-gguf:q4_k_m` | 0.0-0.3 | Deep reasoning |

---

## Engine URLs

```yaml
# llama.cpp (default)
http://localhost:12434/engines/llama.cpp/v1

# Ollama
http://localhost:11434/v1

# LM Studio
http://localhost:1234/v1

# OpenAI Compatible
https://api.openai.com/v1

# HuggingFace Inference
https://api-inference.huggingface.co/models/MODEL_ID

# Cerebras
https://api.cerebras.ai/v1
```

---

## Temperature Guidelines

| Temperature | Behavior | Use Cases |
|-------------|----------|-----------|
| **0.0** | Deterministic, consistent | Tools, structured output, compression, routing |
| **0.1-0.3** | Very focused, minimal variation | Factual Q&A, reasoning |
| **0.4-0.6** | Balanced, reliable | General chat |
| **0.7-0.9** | Creative, varied | Storytelling, brainstorming |
| **1.0+** | Highly random | Experimental, artistic |

---

## Context Management

### History Control

```go
// Stateful (default)
agents.Config{
    KeepConversationHistory: true, // Messages accumulate
}

// Stateless
agents.Config{
    KeepConversationHistory: false, // Only system message persists
}
```

### Context Size Management

```go
// Check size
contextSize := agent.GetContextSize()

// Conditional compression
if contextSize > 8000 {
    compressed, _ := compressorAgent.CompressContextStream(
        agent.GetMessages(),
        streamCallback,
    )
    agent.ResetMessages()
    agent.AddMessage(roles.System, compressed.CompressedText)
}

// Manual history management
messages := agent.GetMessages()
agent.RemoveLastNMessages(3)
agent.ResetMessages()
```

---

## Common Patterns

### Docker/Environment Configuration

```go
// Helper function
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// Usage
engineURL := getEnv("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
modelID := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")

// OR: Use SDK helper
import "github.com/snipwise/nova/nova-sdk/toolbox/env"

engineURL := env.GetEnvOrDefault("ENGINE_URL", "default")
```

### Error Handling

```go
// Development
if err != nil {
    panic(err)
}

// Production
if err != nil {
    log.Printf("Error: %v", err)
    return fmt.Errorf("operation failed: %w", err)
}
```

### Tool Execution Function

```go
func executeFunction(functionName string, arguments string) (string, error) {
    switch functionName {
    case "tool_name":
        var args struct {
            Param1 string `json:"param1"`
        }
        if err := json.Unmarshal([]byte(arguments), &args); err != nil {
            return `{"error": "Invalid arguments"}`, nil
        }

        result := performAction(args.Param1)
        return fmt.Sprintf(`{"result": "%s"}`, result), nil

    default:
        return `{"error": "Unknown function"}`, fmt.Errorf("unknown: %s", functionName)
    }
}
```

---

## Version Requirements

- **Go**: 1.25.4+ (minimum)
- **Nova SDK**: Latest from `github.com/snipwise/nova`

```bash
go get github.com/snipwise/nova@latest
go mod tidy
```
