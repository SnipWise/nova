# Crew Server Agent - Functional Presentation

## Overview

The **Crew Server Agent** is an advanced multi-agent orchestration system from the N.O.V.A. SDK that manages multiple specialized conversational agents and intelligently routes requests to the most appropriate agent based on topic detection.

It is a sophisticated HTTP/REST API component that extends the capabilities of the regular Server Agent by adding intelligent agent orchestration, dynamic routing, and seamless context switching between specialized domain experts.

## Main Features

### 1. Multi-Agent Crew Management

- **Multiple specialized chat agents** managed in a crew
- **Dynamic agent switching** based on conversation topics
- **Runtime crew management**: add or remove agents on the fly
- **Specialized domain expertise**: each agent can be tuned for specific topics
- **Transparent delegation**: appears as a single coherent assistant to users

### 2. Intelligent Topic Detection and Routing

- **Automatic topic classification** using a structured orchestrator agent
- **Custom routing logic** via configurable matching function
- **Seamless agent switching** while maintaining conversation context
- **Fallback mechanism** for unmatched topics
- **Intent-based delegation** for optimal response quality

### 3. REST API with SSE Streaming

- **RESTful HTTP endpoints** for chat completions
- **Real-time streaming** via Server-Sent Events (SSE)
- **Progressive responses** for a smooth user experience
- **Real-time error handling** via SSE

### 4. Tool Calling with Confirmation Workflow

- **Automatic detection** of tool calls in user requests
- **Parallel calls** of multiple tools
- **Web-based validation workflow**: tools require approval before execution
- **Notification system** via SSE to inform the client of pending operations
- **Operation management**: validation, cancellation, reset

### 5. RAG (Retrieval-Augmented Generation)

- **Similarity search** in a vector document database
- **Automatic injection** of relevant context before generation
- **Flexible configuration**: similarity threshold, maximum number of documents
- **Enhanced responses** with external information

### 6. Intelligent Context Compression

- **Automatic management** of token limits
- **On-demand or automatic compression** beyond a threshold
- **Safety thresholds**: warning at 80%, reset at 90%
- **Preservation of essential context** while respecting model limits

### 7. Conversational Memory Management

- **Message history** persisted during the session
- **Token counting** to monitor usage
- **History reset**
- **JSON export** of conversations
- **Listing** of all messages

## Architecture

### Main Structure

```go
type CrewServerAgent struct {
    // Crew Management
    chatAgents       map[string]*chat.Agent // Map of specialized agents
    currentChatAgent *chat.Agent            // Currently active agent

    // Specialized Agents
    toolsAgent        *tools.Agent           // Optional: tool detection/execution
    ragAgent          *rag.Agent             // Optional: document retrieval
    compressorAgent   *compressor.Agent      // Optional: context compression
    orchestratorAgent *structured.Agent[Intent] // Optional: topic detection

    // Orchestration Configuration
    matchAgentIdToTopicFn func(string) string // Custom routing logic

    // RAG Configuration
    similarityLimit float64  // Default: 0.6
    maxSimilarities int      // Default: 3

    // Compression Configuration
    contextSizeLimit int     // Default: 8000

    // Server Configuration
    port string
    ctx  context.Context
    log  logger.Logger

    // Operation Management
    pendingOperations       map[string]*PendingOperation
    operationsMutex         sync.RWMutex
    stopStreamChan          chan bool
    currentNotificationChan chan ToolCallNotification
    notificationChanMutex   sync.Mutex

    // Custom Tool Executor
    executeFn func(string, string) (string, error)
}
```

### Intent Detection Structure

```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}
```

### Specialized Agents

The Crew Server Agent orchestrates five types of agents:

1. **Chat Agents** (required): multiple specialized conversational agents, each with domain expertise
2. **Orchestrator Agent** (optional): structured agent for topic detection and routing
3. **Tools Agent** (optional): function detection and execution
4. **RAG Agent** (optional): relevant document search
5. **Compressor Agent** (optional): context compression

## HTTP Endpoints

### Completion

#### `POST /completion`

Stream a chat completion with intelligent agent routing and SSE.

**Request:**

```json
{
  "data": {
    "message": "Your question here"
  }
}
```

**Response:** SSE stream with text chunks

```
data: {"message": "text chunk"}
data: {"message": "more text"}
data: {"message": "", "finish_reason": "stop"}
```

**Tool notifications (if tools agent configured):**

```json
data: {
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_0x140003dcbe0",
  "message": "Tool call detected: function_name"
}
```

#### `POST /completion/stop`

Stop the current streaming operation.

### Memory Management

#### `POST /memory/reset`

Clear conversation history.

**Response:**

```json
{
  "status": "ok",
  "message": "Memory reset"
}
```

#### `GET /memory/messages/list`

Retrieve all messages from history.

**Response:**

```json
{
  "messages": [
    { "role": "user", "content": "..." },
    { "role": "assistant", "content": "..." }
  ]
}
```

#### `GET /memory/messages/context-size`

Get token count and statistics.

**Response:**

```json
{
  "context_size": 1234,
  "message": "Current context size: 1234 tokens"
}
```

### Tool Operation Management

#### `POST /operation/validate`

Approve a pending tool call.

**Request:**

```json
{
  "operation_id": "op_0x140003dcbe0"
}
```

**Response:**

```json
{
  "status": "validated",
  "operation_id": "op_0x140003dcbe0"
}
```

#### `POST /operation/cancel`

Reject a pending tool call.

**Request:**

```json
{
  "operation_id": "op_0x140003dcbe0"
}
```

**Response:**

```json
{
  "status": "cancelled",
  "operation_id": "op_0x140003dcbe0"
}
```

#### `POST /operation/reset`

Cancel all pending operations.

**Response:**

```json
{
  "status": "ok",
  "message": "All pending operations cancelled"
}
```

### Information

#### `GET /models`

Information about used models.

**Response:**

```json
{
  "chat_models": {
    "coder": "hf.co/menlo/jan-nano-gguf:q4_k_m",
    "cook": "hf.co/menlo/jan-nano-gguf:q4_k_m",
    "thinker": "hf.co/menlo/jan-nano-gguf:q4_k_m"
  },
  "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "embeddings_model": "all-minilm:l6-v2",
  "orchestrator_model": "hf.co/menlo/jan-nano-gguf:q4_k_m"
}
```

#### `GET /health`

Server health check.

**Response:**

```json
{
  "status": "ok",
  "message": "Server is running"
}
```

## Request Execution Flow

### Complete `/completion` Request Processing

1. **Context Compression** (if compressor agent configured)
   - Check context size
   - Compress if threshold exceeded

2. **Request Parsing**
   - JSON decoding of message

3. **SSE Streaming Setup**
   - `text/event-stream` headers
   - Notification channel creation

4. **Tool Detection** (if tools agent configured)
   - Call `DetectParallelToolCallsWithConfirmation()`
   - Send SSE notifications for pending operations
   - Wait for validation/cancellation
   - Execute approved tools
   - Add results to context

5. **RAG Similarity Search** (if RAG agent configured)
   - Search for relevant documents
   - Inject as system message

6. **Topic Detection and Agent Routing** (if orchestrator agent configured)
   - Call `DetectTopicThenGetAgentId()`
   - Analyze user query with orchestrator agent
   - Extract topic from structured response
   - Match topic to agent ID via `matchAgentIdToTopicFn`
   - Switch `currentChatAgent` if different agent matched

7. **Streaming Completion Generation**
   - Stream chunks via SSE using the selected chat agent
   - Handle stop signals
   - Send finish reason

8. **Error Handling**
   - Stream errors via SSE

## Usage Examples

### 1. Basic Crew Server Agent (Multiple Specialized Agents)

```go
import (
    "context"
    "log"
    "strings"

    "nova-sdk/agents"
    "nova-sdk/agents/chat"
    "nova-sdk/agents/crewserver"
    "nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Create specialized chat agents
    coderAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "coder",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are an expert programmer specialized in Go, Python, and web development.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.3),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    cookAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "cook",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a professional chef specialized in world cuisine and cooking techniques.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.7),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    thinkerAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "thinker",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a philosopher and psychologist specialized in critical thinking and self-improvement.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.6),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create agent crew
    agentCrew := map[string]*chat.Agent{
        "coder":   coderAgent,
        "cook":    cookAgent,
        "thinker": thinkerAgent,
    }

    // Define routing logic
    matchAgentFunction := func(topic string) string {
        switch strings.ToLower(topic) {
        case "coding", "programming", "development", "software":
            return "coder"
        case "cooking", "recipe", "food", "cuisine":
            return "cook"
        case "philosophy", "psychology", "thinking", "mindfulness":
            return "thinker"
        default:
            return "coder" // Default agent
        }
    }

    // Create crew server agent
    crewAgent, err := crewserver.NewAgent(
        ctx,
        agentCrew,
        "coder", // Initial active agent
        ":8080",
        matchAgentFunction,
        nil, // No custom execute function
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(crewAgent.StartServer())
}
```

**Usage:**

```bash
# Programming question - routed to coder agent
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "How to use switch case in Golang?"}}'

# Cooking question - routed to cook agent
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "What is a recipe for Hawaiian pizza?"}}'

# Psychology question - routed to thinker agent
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "How to manage anxiety?"}}'
```

### 2. Crew Server Agent with Intelligent Orchestration

```go
import (
    "nova-sdk/agents/structured"
)

func main() {
    ctx := context.Background()

    // Create agent crew (same as above)
    agentCrew := map[string]*chat.Agent{
        "coder":   coderAgent,
        "cook":    cookAgent,
        "thinker": thinkerAgent,
    }

    // Create crew server agent
    crewAgent, err := crewserver.NewAgent(
        ctx,
        agentCrew,
        "coder",
        ":8080",
        matchAgentFunction,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create orchestrator agent for topic detection
    orchestratorAgent, err := structured.NewAgent[crewserver.Intent](
        ctx,
        agents.Config{
            Name:      "orchestrator",
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: `You are a topic classifier.
Analyze the user query and determine the main topic.
Possible topics: programming, cooking, philosophy, psychology, general.`,
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Attach orchestrator for intelligent routing
    crewAgent.SetOrchestratorAgent(orchestratorAgent)

    log.Fatal(crewAgent.StartServer())
}
```

**How it works:**

1. User sends: "What's the best way to implement a REST API?"
2. Orchestrator detects topic: "programming"
3. `matchAgentFunction` maps "programming" to "coder"
4. Request is routed to the coder agent
5. Coder agent generates the response

### 3. Full-Featured Crew Server Agent (Orchestration + Tools + RAG + Compression)

```go
import (
    "nova-sdk/agents/tools"
    "nova-sdk/agents/rag"
    "nova-sdk/agents/compressor"
    "nova-sdk/chunks"
)

func executeFunction(functionName, arguments string) (string, error) {
    // Your tool implementations
    switch functionName {
    case "calculator":
        // Implement calculator logic
        return "42", nil
    default:
        return "", fmt.Errorf("unknown function: %s", functionName)
    }
}

func main() {
    ctx := context.Background()

    // Create crew (same as above)
    agentCrew := map[string]*chat.Agent{
        "coder":   coderAgent,
        "cook":    cookAgent,
        "thinker": thinkerAgent,
    }

    // Create crew server agent
    crewAgent, err := crewserver.NewAgent(
        ctx,
        agentCrew,
        "coder",
        ":8080",
        matchAgentFunction,
        executeFunction,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Add orchestrator
    orchestratorAgent, err := structured.NewAgent[crewserver.Intent](
        ctx,
        orchestratorConfig,
        orchestratorModelConfig,
    )
    crewAgent.SetOrchestratorAgent(orchestratorAgent)

    // Add tools agent
    toolsAgent, err := tools.NewAgent(
        ctx,
        toolsConfig,
        models.Config{
            Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature:       models.Float64(0.0),
            ParallelToolCalls: models.Bool(true),
        },
        tools.WithTools(GetToolsIndex()),
    )
    if err != nil {
        log.Fatal(err)
    }
    crewAgent.SetToolsAgent(toolsAgent)

    // Add RAG agent
    ragAgent, err := rag.NewAgent(ctx, ragConfig, embeddingsModelConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Load documents
    documents := []string{
        "Content of document 1...",
        "Content of document 2...",
    }

    for _, content := range documents {
        chunks := chunks.SplitMarkdownBySections(content)
        for _, chunk := range chunks {
            ragAgent.SaveEmbedding(chunk)
        }
    }
    crewAgent.SetRagAgent(ragAgent)
    crewAgent.SetSimilarityLimit(0.6)
    crewAgent.SetMaxSimilarities(3)

    // Add compressor agent
    compressorAgent, err := compressor.NewAgent(
        ctx,
        compressorConfig,
        compressorModelConfig,
        compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    )
    if err != nil {
        log.Fatal(err)
    }
    crewAgent.SetCompressorAgent(compressorAgent)
    crewAgent.SetContextSizeLimit(3000)

    log.Fatal(crewAgent.StartServer())
}
```

## Configuration Methods

### Crew Management

```go
agent.AddChatAgentToCrew("expert", expertAgent)    // Add new agent
agent.RemoveChatAgentFromCrew("expert")            // Remove agent
agents := agent.GetChatAgents()                    // Get all agents
agent.SetChatAgents(newAgentMap)                   // Replace all agents
```

### Server Configuration

```go
agent.SetPort(":8080")
port := agent.GetPort()
```

### Agent Configuration

```go
agent.SetOrchestratorAgent(orchestratorAgent)  // Enable intelligent routing
agent.SetToolsAgent(toolsAgent)
agent.SetRagAgent(ragAgent)
agent.SetCompressorAgent(compressorAgent)
```

### RAG Configuration

```go
agent.SetSimilarityLimit(0.6)      // Similarity threshold (0.0 - 1.0)
agent.SetMaxSimilarities(3)        // Max number of documents to retrieve
```

### Compression Configuration

```go
agent.SetContextSizeLimit(8000)    // Token limit
agent.CompressChatAgentContext()   // Force compression
agent.CompressChatAgentContextIfOverLimit() // Conditional compression
```

### Memory Management

```go
agent.ResetMessages()              // Clear history
agent.AddMessage(role, content)    // Add a message
messages := agent.GetMessages()    // Retrieve all messages
size := agent.GetContextSize()     // Get context size
json := agent.ExportMessagesToJSON() // Export to JSON
```

### Custom Execute Function

```go
agent.SetExecuteFunction(func(functionName, arguments string) (string, error) {
    // Your execution logic
    return result, nil
})
```

## SSE Message Format

### Regular text chunk

```
data: {"message": "text chunk"}
```

### Tool call notification

```json
data: {
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_0x140003dcbe0",
  "function_name": "calculator",
  "arguments": "{\"a\":40,\"b\":2}",
  "message": "Tool call detected: calculator"
}
```

### Finish response

```
data: {"message": "", "finish_reason": "stop"}
```

### Error

```
data: {"error": "error message"}
```

## Topic Detection and Routing

### How It Works

The orchestration system uses a two-step process:

1. **Topic Detection**: The orchestrator agent analyzes the user query and determines the main topic
2. **Agent Matching**: The `matchAgentIdToTopicFn` maps the detected topic to a specific agent ID

### Example Routing Configuration

```go
matchAgentFunction := func(topic string) string {
    topicLower := strings.ToLower(topic)

    // Programming-related topics
    if strings.Contains(topicLower, "coding") ||
       strings.Contains(topicLower, "programming") ||
       strings.Contains(topicLower, "software") ||
       strings.Contains(topicLower, "development") {
        return "coder"
    }

    // Cooking-related topics
    if strings.Contains(topicLower, "cooking") ||
       strings.Contains(topicLower, "recipe") ||
       strings.Contains(topicLower, "food") ||
       strings.Contains(topicLower, "cuisine") {
        return "cook"
    }

    // Philosophy/psychology topics
    if strings.Contains(topicLower, "philosophy") ||
       strings.Contains(topicLower, "psychology") ||
       strings.Contains(topicLower, "thinking") ||
       strings.Contains(topicLower, "mindfulness") {
        return "thinker"
    }

    // Default fallback
    return "coder"
}
```

### Agent Switching Rules

- The system can switch agents between requests
- Agent switching is seamless and maintains conversation context
- Cannot remove the currently active agent from the crew
- Each agent has its own personality and system instructions

## Security and Concurrency

### Thread-safety

- **Mutex-protected operations map** for concurrent tool calls
- **Channel-based communication** for operation responses
- **Notification channel locking** to prevent race conditions
- **Safe crew management** with concurrent request handling

### Operation Management

- Each tool call receives a unique ID
- Pending operations are stored in a thread-safe map
- Response channels enable asynchronous communication
- Timeouts and cancellations handled properly

## Logging

The Crew Server Agent uses the N.O.V.A. SDK logging system:

```bash
# Control via environment variable
export NOVA_LOG_LEVEL=DEBUG  # DEBUG, INFO, WARN, ERROR
```

## Default Values

| Parameter           | Default Value   | Description                        |
| ------------------- | --------------- | ---------------------------------- |
| Port                | Set at creation | HTTP server port                   |
| Similarity Limit    | 0.6             | RAG similarity threshold (0.0-1.0) |
| Max Similarities    | 3               | Max number of RAG documents        |
| Context Size Limit  | 8000            | Token limit before compression     |
| Compression Warning | 80%             | Warning for approaching limit      |
| Compression Reset   | 90%             | Forced reset                       |

## Test Scripts

The directory contains several test scripts:

- **`01-programming-stream.sh`** - Programming question test (routed to coder)
- **`02-cooking-stream.sh`** - Cooking question test (routed to cook)
- **`03-psychology-stream.sh`** - Psychology question test (routed to thinker)
- **`call-tool.sh`** - Trigger tool calls
- **`validate.sh`** - Approve pending operation
- **`cancel.sh`** - Reject pending operation
- **`reset.sh`** - Cancel all operations

## Sample Directory

Complete examples are available in `/samples`:

- **`55-crew-server-agent`** - Full crew server agent with orchestration

## Comparison: Server Agent vs Crew Server Agent

| Feature              | Server Agent           | Crew Server Agent                  |
| -------------------- | ---------------------- | ---------------------------------- |
| **Chat Agents**      | Single agent           | Multiple specialized agents        |
| **Agent Switching**  | No                     | Yes - dynamic based on topic       |
| **Orchestration**    | No                     | Yes - with structured agent        |
| **Topic Detection**  | No                     | Yes - automatic routing            |
| **Agent Management** | Static                 | Dynamic add/remove agents          |
| **Routing Logic**    | N/A                    | Customizable matching function     |
| **Use Case**         | Single-purpose chatbot | Multi-domain intelligent assistant |
| **Tool Calling**     | Yes                    | Yes                                |
| **RAG**              | Yes                    | Yes                                |
| **Compression**      | Yes                    | Yes                                |
| **SSE Streaming**    | Yes                    | Yes                                |

## Advantages of Crew Orchestration

1. **Specialized Expertise**: Each agent can be optimized for specific domains
2. **Better Response Quality**: Domain-specific system instructions lead to more accurate answers
3. **Scalability**: Easy to add new specialized agents
4. **Flexibility**: Dynamic routing based on conversation context
5. **Maintainability**: Separate concerns by domain
6. **Performance**: Different temperature settings for different tasks

## Conclusion

The **Crew Server Agent** is a powerful multi-agent orchestration system from the N.O.V.A. SDK that transforms multiple specialized conversational agents into a unified intelligent assistant with:

- ✅ Intelligent topic detection and routing
- ✅ Multiple specialized domain experts
- ✅ Real-time streaming via SSE
- ✅ Tool calling with validation workflow
- ✅ Relevant document retrieval (RAG)
- ✅ Intelligent context compression
- ✅ Complete conversational memory management
- ✅ Dynamic crew management
- ✅ Modular and extensible architecture

It enables rapid deployment of sophisticated multi-domain AI assistants capable of handling diverse user requests by automatically delegating to specialized agents, while maintaining fine-grained control over sensitive operations like tool execution.
