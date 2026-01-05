# Orchestrator Agent

## Description

The **Orchestrator Agent** is a specialized agent for topic/intent detection from user messages. It uses a structured agent internally to generate output in `agents.Intent` format containing the identified discussion topic.

## Features

- **Topic detection** : Identifies the main topic of a conversation
- **Intent detection** : Extracts user intent from messages
- **Structured output** : Returns an `Intent` object with the `TopicDiscussion` field
- **Intelligent routing** : Used by Crew Agents to route requests to the appropriate agent

## Use cases

The Orchestrator Agent is primarily used to:
- **Route requests** in multi-agent systems (Crew Agent)
- **Classify questions** by topic
- **Detect user intent** to trigger specific actions

## Creating an Orchestrator Agent

### Basic syntax

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Agent configuration
agentConfig := agents.Config{
    Name: "Orchestrator",
    Instructions: `You are a topic detection assistant. Analyze user messages
and identify the main topic of discussion. Return only the topic category.`,
}

// Model configuration
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    ModelID: "qwen2.5:1.5b", // Fast model for classification
}

// Create the agent
agent, err := orchestrator.NewAgent(ctx, agentConfig, modelConfig)
if err != nil {
    log.Fatal(err)
}
```

## Intent structure

The agent returns an `agents.Intent` object:

```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}
```

## Main methods

### IdentifyIntent - Detection from messages

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Identify intent from messages
userMessages := []messages.Message{
    {
        Role:    roles.User,
        Content: "How to make a Neapolitan pizza?",
    },
}

intent, finishReason, err := agent.IdentifyIntent(userMessages)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Topic:", intent.TopicDiscussion) // "cooking" or "food"
fmt.Println("Finish reason:", finishReason)   // "stop"
```

### IdentifyTopicFromText - Detection from text

```go
// Convenient method to detect topic from simple text
topic, err := agent.IdentifyTopicFromText("Explain the Factory pattern in Go")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Detected topic:", topic) // "programming" or "coding"
```

### Message management

```go
// Add a message
agent.AddMessage(roles.User, "Question...")

// Add multiple messages
messages := []messages.Message{
    {Role: roles.User, Content: "Question 1"},
    {Role: roles.Assistant, Content: "Answer 1"},
}
agent.AddMessages(messages)

// Get all messages
allMessages := agent.GetMessages()

// Reset messages
agent.ResetMessages()
```

### Getters and Setters

```go
// Configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)

// Information
name := agent.GetName()
modelID := agent.GetModelID()
kind := agent.Kind() // Returns agents.Orchestrator

// Context
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requests/Responses (debugging)
lastRequestJSON, _ := agent.GetLastRequestJSON()
lastResponseJSON, _ := agent.GetLastResponseJSON()
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
```

## Usage with Crew Agent

The Orchestrator Agent is designed to be used with Crew Agent for automatic request routing:

```go
// Create the orchestrator
orchestratorAgent, _ := orchestrator.NewAgent(ctx, orchestratorConfig, modelConfig)

// Topic → agent ID mapping function
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

// Create crew with orchestrator
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(matchAgentFn),
)

// Orchestrator automatically detects topic and routes to appropriate agent
result, _ := crewAgent.StreamCompletion("How to make carbonara?", callback)
// → Orchestrator detects "cooking" → routes to "cook" agent
```

## Complete example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration with topic detection instructions
    agentConfig := agents.Config{
        Name: "TopicDetector",
        Instructions: `You are a topic classification assistant.
Analyze the user's message and identify the main topic category.

Categories:
- coding/programming: Questions about software development, code, debugging
- cooking/food: Questions about recipes, cooking techniques, ingredients
- philosophy/thinking: Questions about ideas, concepts, psychology
- generic: Everything else

Return only the topic category.`,
    }

    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        ModelID:   "qwen2.5:1.5b",
    }

    // Create the orchestrator
    agent, err := orchestrator.NewAgent(ctx, agentConfig, modelConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Detection examples
    questions := []string{
        "How to implement a singleton in Go?",
        "What's the recipe for margherita pizza?",
        "What is free will?",
    }

    for _, question := range questions {
        topic, err := agent.IdentifyTopicFromText(question)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        fmt.Printf("Question: %s\n", question)
        fmt.Printf("→ Topic: %s\n\n", topic)
    }

    // Expected output:
    // Question: How to implement a singleton in Go?
    // → Topic: coding
    //
    // Question: What's the recipe for margherita pizza?
    // → Topic: cooking
    //
    // Question: What is free will?
    // → Topic: philosophy
}
```

## OrchestratorAgent interface

The Orchestrator Agent implements the `agents.OrchestratorAgent` interface:

```go
type OrchestratorAgent interface {
    // IdentifyIntent sends messages and returns the identified intent
    IdentifyIntent(userMessages []messages.Message) (intent *Intent, finishReason string, err error)

    // IdentifyTopicFromText is a convenience method that takes a text string and returns the topic
    IdentifyTopicFromText(text string) (string, error)
}
```

## Internal architecture

The Orchestrator Agent uses internally a `structured.Agent[agents.Intent]` which:
1. Takes user messages
2. Uses the LLM model to generate structured output
3. Parses the JSON output to `agents.Intent` object
4. Returns the `TopicDiscussion` field

## Notes

- **Kind** : Returns `agents.Orchestrator`
- **Based on Structured Agent** : Uses `structured.Agent[agents.Intent]` internally
- **Structured output** : Guarantees consistent JSON format with `topic_discussion` field
- **Empty error** : Returns an error if `userMessages` is empty
- **Recommended model** : Use a fast model (e.g., `qwen2.5:1.5b`) for quick classification
- **Critical instructions** : Agent instructions must guide the model to correctly identify topics

## Recommendations

### Effective instructions

```go
Instructions: `You are a topic classifier. Analyze the message and return ONE topic category.

Categories:
- coding: programming, software, code
- cooking: recipes, food, ingredients
- philosophy: ideas, concepts, thinking
- science: physics, chemistry, biology
- generic: everything else

Return only the category name.`
```

### Appropriate model

- **Fast model** : `qwen2.5:1.5b`, `lucy`, `jan-nano` for quick classification
- **Avoid large models** : Topic detection doesn't require heavy models

### Optimal usage

- **Crew routing** : Use with `crew.Agent` for automatic routing
- **Clear instructions** : Clearly define topic categories
- **Explicit mapping** : Use `WithMatchAgentIdToTopicFn` to map topics to agents
