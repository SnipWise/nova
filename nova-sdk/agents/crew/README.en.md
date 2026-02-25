# Crew Agent

## Description

The **Crew Agent** is a composite agent that orchestrates multiple chat agents (`chat.Agent`) to answer complex questions. It can intelligently route requests to the most appropriate agent and combine multiple capabilities (Tools, RAG, Compressor, Orchestrator).

## Features

- **Multi-agent** : Manages multiple chat agents with dynamic routing
- **Orchestration** : Topic/intent detection and automatic routing to the appropriate agent
- **Tools Agent** : Function calling with user confirmation
- **RAG Agent** : Similarity search and context enrichment
- **Compressor Agent** : Automatic context compression when limit is reached
- **Human-in-the-loop** : Customizable function call validation

## Creating a Crew Agent

### Syntax with options

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents/crew"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
)

// Create with a single agent
agent, err := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
)

// Create with multiple agents
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "writer":  writerAgent,
}

agent, err := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "coder"), // "coder" is the default agent
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithToolsAgent(toolsAgent),
    crew.WithRagAgent(ragAgent),
    crew.WithCompressorAgentAndContextSize(compressorAgent, 8000),
    crew.WithExecuteFn(myCustomExecutor),
    crew.WithConfirmationPromptFn(myConfirmationPrompt),
)
```

### Available options

| Option | Description |
|--------|-------------|
| `WithSingleAgent(chatAgent)` | Creates a crew with a single agent (ID: "single") |
| `WithAgentCrew(agentCrew, selectedAgentId)` | Defines multiple agents with the initial selected agent |
| `WithMatchAgentIdToTopicFn(fn)` | Custom function to map topic → agent ID |
| `WithOrchestratorAgent(orchestratorAgent)` | Agent for topic/intent detection |
| `WithExecuteFn(fn)` | Custom function executor for tools |
| `WithConfirmationPromptFn(fn)` | Custom confirmation function for tool calls |
| `WithToolsAgent(toolsAgent)` | Adds an agent for function execution |
| `WithTasksAgent(tasksAgent)` | Adds a tasks agent for task planning and orchestration |
| `WithRagAgent(ragAgent)` | Adds a RAG agent for document retrieval |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | RAG with similarity configuration |
| `WithCompressorAgent(compressorAgent)` | Adds an agent for context compression |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Compressor with context size limit |

## Main methods

### Crew management

```go
// Get all agents
chatAgents := agent.GetChatAgents()

// Add an agent to the crew
err := agent.AddChatAgentToCrew("expert", expertAgent)

// Remove an agent from the crew (except the active agent)
err := agent.RemoveChatAgentFromCrew("expert")

// Get/Set the active agent
currentId := agent.GetSelectedAgentId()
err := agent.SetSelectedAgentId("thinker")
```

### Completion generation

```go
// Streaming with callback
result, err := agent.StreamCompletion(question, func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
})
```

### Orchestrator management

```go
// Set the orchestrator agent
agent.SetOrchestratorAgent(orchestratorAgent)

// Detect topic and get appropriate agent ID
agentId, err := agent.DetectTopicThenGetAgentId("How to cook a pizza?")
// → Returns "cook" if matchAgentIdToTopicFn maps it
```

### Auxiliary agents management

```go
// Tools Agent
agent.SetToolsAgent(toolsAgent)
agent.SetExecuteFunction(myExecutor)
agent.SetConfirmationPromptFunction(myConfirmationFn)

// RAG Agent
agent.SetRagAgent(ragAgent)

// Compressor Agent
agent.SetCompressorAgent(compressorAgent)
```

### Methods inherited from chat.Agent

```go
// Messages
agent.GetMessages()
agent.AddMessage(roles.User, "Question...")
agent.ResetMessages()

// Context
contextSize := agent.GetContextSize()

// Generation (delegates to active agent)
agent.GenerateCompletion(messages)
agent.GenerateStreamCompletion(messages, callback)
agent.GenerateCompletionWithReasoning(messages)
agent.GenerateStreamCompletionWithReasoning(messages, reasoningCb, responseCb)

// Export
jsonData, err := agent.ExportMessagesToJSON()
```

## Processing pipeline (StreamCompletion)

```
1. Context compression (if CompressorAgent configured)
   ↓
2. Tool call detection (if ToolsAgent configured)
   ↓
3. User confirmation request (via confirmationPromptFn)
   ↓
4. Function execution (if confirmed)
   ↓
5. Add result to context
   ↓
6. RAG search (if RagAgent configured)
   ↓
7. Topic detection and routing (if OrchestratorAgent configured)
   ↓
8. Response generation with the appropriate agent
   ↓
9. State cleanup
```

## Intelligent routing with Orchestrator

The orchestrator detects the topic of the question and routes to the appropriate agent.

### Configuring topic → agent mapping

```go
// Custom mapping function
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding", "programming", "development":
        return "coder"
    case "philosophy", "thinking", "psychology":
        return "thinker"
    case "cooking", "recipe", "food":
        return "cook"
    default:
        return "generic"
    }
}

agent, err := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(matchAgentFn),
)
```

### Automatic detection during StreamCompletion

When `StreamCompletion` is called and an orchestrator is configured:
1. The orchestrator detects the topic of the question
2. The `matchAgentIdToTopicFn` function maps the topic to an agent ID
3. The active agent (`currentChatAgent`) is automatically switched
4. The response is generated by the newly selected agent

```go
// User asks a question about cooking
result, err := agent.StreamCompletion("How to make a pizza?", callback)
// → Orchestrator detects "cooking" → routes to "cook"
```

## Complete example

```go
ctx := context.Background()

// Create multiple specialized agents
coderAgent, _ := chat.NewAgent(ctx,
    agents.Config{Name: "Coder", Instructions: "Expert in programming"},
    modelConfig,
)
thinkerAgent, _ := chat.NewAgent(ctx,
    agents.Config{Name: "Thinker", Instructions: "Expert in philosophy"},
    modelConfig,
)
cookAgent, _ := chat.NewAgent(ctx,
    agents.Config{Name: "Cook", Instructions: "Expert in cooking"},
    modelConfig,
)

// Create the orchestrator
orchestratorAgent, _ := orchestrator.NewAgent(ctx, agentConfig, modelConfig)

// Topic → agent mapping function
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding", "programming":
        return "coder"
    case "philosophy", "thinking":
        return "thinker"
    case "cooking", "food":
        return "cook"
    default:
        return "coder" // Default agent
    }
}

// Create the crew agent
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "cook":    cookAgent,
}

crewAgent, err := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "coder"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(matchAgentFn),
    crew.WithToolsAgent(toolsAgent),
)

// Usage
result, err := crewAgent.StreamCompletion("Explain the Factory pattern", func(chunk, reason string) error {
    fmt.Print(chunk)
    return nil
})
// → Routed to "coder" automatically

result, err = crewAgent.StreamCompletion("What's the recipe for carbonara?", callback)
// → Routed to "cook" automatically
```

## Dynamic crew management

```go
// Add a new agent during execution
expertAgent, _ := chat.NewAgent(ctx, expertConfig, modelConfig)
err := crewAgent.AddChatAgentToCrew("expert", expertAgent)

// Manually switch to an agent
err = crewAgent.SetSelectedAgentId("expert")

// Remove an agent (impossible if it's the active agent)
err = crewAgent.RemoveChatAgentFromCrew("expert")
```

## Notes

- **At least one agent required** : `WithAgentCrew` or `WithSingleAgent` is mandatory
- **Active agent** : Only one agent is active at a time (`currentChatAgent`)
- **Automatic routing** : The orchestrator automatically changes the active agent during `StreamCompletion`
- **Default values** :
  - `similarityLimit`: 0.6
  - `maxSimilarities`: 3
  - `contextSizeLimit`: 8000
- **Kind** : Returns `agents.Composite`
- **Delegated methods** : `GetName()`, `GetModelID()`, etc. are delegated to `currentChatAgent`

## Legacy constructor

A simplified constructor also exists (without options):

```go
agent, err := crew.NewSimpleAgent(ctx, agentCrew, "coder")
```

**Note** : Prefer `NewAgent` with options for more flexibility.
