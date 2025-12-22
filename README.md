# Nova
**N**eural **O**ptimized **V**irtual **A**ssistant
> Composable AI agents framework in Go

## Nova SDK - Getting Started Examples

> Nova SDK has been tested with:
> - [Docker Model Runner](https://docs.docker.com/ai/model-runner/)
> - [Ollama](https://ollama.com/)

> This `README.md` file is a work in progress and will be expanded with more examples soon.

### Installation

```bash
go get github.com/snipwise/nova@latest
```


### Chat agent
> Simple completion

> `go test -v -run TestSimpleChatAgent ./getting-started/tests` 
```golang
package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func TestSimpleChatAgent(t *testing.T) {

	ctx := context.Background()

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
			MaxTokens:   models.Int(2000),
		},
	)
	if err != nil {
		panic(err)
	}

	display := func(result *chat.CompletionResult) {
		fmt.Println()
		fmt.Println("Response:\n", result.Response)
		fmt.Println()
		fmt.Println("Finish reason:\n", result.FinishReason)
		fmt.Println(strings.Repeat("-", 40))
	}

	// Simple chat using only Message structs
	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] who is James T Kirk?"},
	})

	if err != nil {
		panic(err)
	}
	display(result)

	// Context is maintained automatically
	// Continue the conversation
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] who is his best friend?"},
	})

	if err != nil {
		panic(err)
	}
	display(result)

}
```

### Chat agent with streaming
> Simple streaming completion

`go test -v -run TestSimpleStreamChatAgent ./getting-started/tests` 
```golang
package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func TestSimpleStreamChatAgent(t *testing.T) {
	ctx := context.Background()

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		panic(err)
	}

	displayContextSize := func(agent *chat.Agent, result *chat.CompletionResult) {
		fmt.Println()
		fmt.Println("Finish reason:\n", result.FinishReason)
		fmt.Printf("Context size: %d characters\n", agent.GetContextSize())
		fmt.Println(strings.Repeat("-", 40))
	}

	// Chat with streaming
	result, err := agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is James T Kirk?"},
		},
		func(chunk string, finishReason string) error {
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	displayContextSize(agent, result)

	result, err = agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is his best friend?"},
		},
		func(chunk string, finishReason string) error {
			// Simple callback that receives strings only
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	displayContextSize(agent, result)
}
```

### RAG agent (Retrieval-Augmented Generation)
> In memory vector store with simple RAG agent

**Create embeddings, store them in memory, and query them.**:

`go test -v -run TestRagAgent ./getting-started/tests` 
```golang
package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/models"
)

func TestRagAgent(t *testing.T) {
	ctx := context.Background()

	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "ai/mxbai-embed-large",
		},		
	)
	if err != nil {
		panic(err)
	}

	txtChunks := []string{
		"Squirrels run in the forest",
		"Birds fly in the sky",
		"Frogs swim in the pond",
		"Fishes swim in the sea",
		"Lions roar in the savannah",
		"Eagles soar above the mountains",
		"Dolphins leap out of the ocean",
		"Bears fish in the river",
	}
	for _, chunk := range txtChunks {
		err := agent.SaveEmbedding(chunk)
		if err != nil {
			panic(err)
		}
	}

	query := "Which animals swim?"

	similarities, err := agent.SearchSimilar(query, 0.6)

	if err != nil {
		panic(err)
	}

	fmt.Println("Similarities for query:", query)
	for _, sim := range similarities {
		fmt.Println("Content:", sim.Prompt)
		fmt.Println("Score:", sim.Similarity)
	}
}
```
> 🚧 more kind of RAG agents are coming

### Tools agent
> Agent with tool calling capabilities

**Create an agent that can call multiple tools to complete tasks.**:

`go test -v -run TestToolCompletionAgent ./getting-started/tests`
```golang
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func TestToolCompletionAgent(t *testing.T) {
	ctx := context.Background()

	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},

		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},

		tools.WithTools(getToolsIndex()),
	)

	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	messages := []messages.Message{
		{
			Content: `
			Make the sum of 40 and 2,
			then say hello to Bob and to Sam,
			make the sum of 5 and 37
			Say hello to Alice
			`,
			Role: roles.User,
		},
	}

	result, err := agent.DetectToolCallsLoop(messages, executeFunction)
	if err != nil {
		t.Fatalf("DetectToolCallsLoop failed: %v", err)
	}

	// Display results
	fmt.Println("Finish Reason:", result.FinishReason)
	for _, value := range result.Results {
		fmt.Println("Result for tool:", value)
	}
	fmt.Println("Assistant Message:", result.LastAssistantMessage)

	// Verify we got some results
	if len(result.Results) == 0 {
		t.Error("Expected at least one tool result")
	}
}

func getToolsIndex() []*tools.Tool {
	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	fmt.Printf("🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s!🙂", args.Name)
		return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	case "calculate_sum":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for calculate_sum"}`, nil
		}
		sum := args.A + args.B
		return fmt.Sprintf(`{"result": %g}`, sum), nil

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
```