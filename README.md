# Nova
**N**eural **O**ptimized **V**irtual **A**ssistant
> Composable AI agents framework in Go

Nova specializes in developing generative text AI apps with local tiny language models.

## Introducing Nova: A Go Framework for Local AI Agents"

Nova was developed with one main goal: to create AI agents simply, and above all with local language models, primarily with tiny language models (I like working with models ranging from 0.5B to 8B parameters).
I started developing Nova because I couldn't find a library or framework in Go that suited my needs for developing generative AI applications: lack of features, use of outdated versions of the OpenAI Go SDK (Nova uses OpenAI Go SDK v3), etc.
My preferred "LLM engine" is **[Docker Model Runner](https://docs.docker.com/ai/model-runner/)** used in conjunction with **[Docker Agentic Compose](https://docs.docker.com/ai/compose/models-and-compose/)**, but it's entirely possible to use Nova with other engines, such as Ollama, LM Studio, the Hugging Face API, Cerebras, and others.

> My mentor used to say: *"Always start by showing code"*
```golang
agent, err := chat.NewAgent(
	ctx,
	agents.Config{
		EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
		SystemInstructions: "You are Bob, a helpful AI assistant.",
	},
	models.Config{
		Name:        "ai/qwen2.5:1.5B-F16",
		Temperature: models.Float64(0.8),
	},
)

result, err := agent.GenerateStreamCompletion(
	[]messages.Message{
		{Role: roles.User, Content: "Why is Hawaiian pizza the best?"},
	},
	func(chunk string, finishReason string) error {
		if chunk != "" {
			fmt.Print(chunk)
		}
		return nil
	},
)
```

## Out-of-the-box AI agents

Nova ships with pre-built AI agents that you can compose to create new ones:
- **Chat Agent**: Conversational agent with context management and streaming support.
- **RAG Agent**: Retrieval-Augmented Generation agent with in-memory vector store.
- **Tools Agent**: Agent with tool-calling capabilities, including parallel tool execution and human-in-the-loop confirmation.
- **Compressor Agent**: Agent with context compression for long conversations.
- **Structured Output Agent**: Agent that produces structured outputs using Go structs.
- **Orchestrator Agent**: Specialized agent for topic detection and query routing.
- **Crew Agent**: Multi-agent collaboration framework for complex tasks.
- **Server Agent**: HTTP/REST API server agent with SSE streaming, tool calling, RAG, and context compression.
- **Remote Agent**: Client agent that connects to a Server Agent for distributed AI applications.
- **Crew Agent Server**: Multi-agent server for collaborative AI tasks over HTTP.

> More agents and features will be added soon!

## OpenAI API Compliance

Nova SDK is fully compatible with OpenAI API specifications. 

> Nova SDK has been tested with:
> - **Primarily** [Docker Model Runner](https://docs.docker.com/ai/model-runner/)
> - [Ollama](https://ollama.com/)
> - [LM Studio](https://lmstudio.ai/)
> - [Hugging Face Inference API](https://huggingface.co/inference-api)
> - [Cerebras API](https://inference-docs.cerebras.ai/introduction)

## Nova SDK - Getting Started Examples

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

### Tools agent with parallel tool calls
> Agent with parallel tool calling capabilities

**Create an agent that can execute multiple tools in parallel to complete tasks more efficiently.**:

`go test -v -run TestParallelToolCallsAgent ./getting-started/tests`
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

func TestParallelToolCallsAgent(t *testing.T) {
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
			ParallelToolCalls: models.Bool(true), // IMPORTANT: Enable parallel tool calls
		},

		tools.WithTools(getParallelToolsIndex()),
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

	result, err := agent.DetectParallelToolCalls(messages, executeParallelFunction)
	if err != nil {
		t.Fatalf("DetectParallelToolCalls failed: %v", err)
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

func getParallelToolsIndex() []*tools.Tool {
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

func executeParallelFunction(functionName string, arguments string) (string, error) {
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

### Tools agent with parallel tool calls and confirmation
> Agent with parallel tool calling capabilities and human-in-the-loop confirmation

**Create an agent that can execute multiple tools in parallel with confirmation before execution.**:

`go test -v -run TestParallelToolCallsWithConfirmationAgent ./getting-started/tests`
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

func TestParallelToolCallsWithConfirmationAgent(t *testing.T) {
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
			ParallelToolCalls: models.Bool(true), // IMPORTANT: Enable parallel tool calls
		},

		tools.WithTools(getConfirmationToolsIndex()),
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

	result, err := agent.DetectParallelToolCallsWithConfirmation(
		messages,
		executeFunction,
		confirmationPrompt)

	if err != nil {
		t.Fatalf("DetectParallelToolCallsWithConfirmation failed: %v", err)
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

func confirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	fmt.Printf("🟢 Detected function: %s with arguments: %s\n", functionName, arguments)

	// For automated testing, we auto-approve all tool calls
	// In a real application, you would use prompt.HumanConfirmation here
	fmt.Printf("Auto-approving execution of %s with %s\n", functionName, arguments)
	return tools.Confirmed
}

func getConfirmationToolsIndex() []*tools.Tool {
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

#### Real-world confirmation example

In a production environment, you can implement human confirmation using Go's standard packages:

```golang
import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
)

func confirmationPromptWithHumanInteraction(functionName string, arguments string) tools.ConfirmationResponse {
	fmt.Printf("🟢 Detected function: %s with arguments: %s\n", functionName, arguments)
	fmt.Printf("Execute %s? (y/n/q): ", functionName)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "y", "yes":
		return tools.Confirmed
	case "n", "no":
		return tools.Denied
	case "q", "quit":
		return tools.Quit
	default:
		return tools.Denied
	}
}
```

The confirmation function allows the user to:
- Enter `y` or `yes` to **confirm** the tool execution (returns `tools.Confirmed`)
- Enter `n` or `no` to **deny** the tool execution (returns `tools.Denied`)
- Enter `q` or `quit` to **quit** the entire tool execution loop (returns `tools.Quit`)

## Base Agents

The **Chat**, **Rag**, **Tool** and **Compressor** Nova agents are built on top of **base agents**, which provide the foundational building blocks for AI interactions. Understanding base agents helps you create custom agents or extend existing ones.

### What are Base Agents?

Base agents are the low-level components that handle direct communication with LLM engines using the **OpenAI API**. Nova provides two primary base agents:

1. **Completion Agent** - For non-streaming text generation
2. **Stream Completion Agent** - For streaming text generation with real-time responses


### When to Use Base Agents

Use base agents directly when you need:

- **Maximum control** over the LLM interaction
- **Custom agent behavior** not covered by high-level agents
- **Minimal overhead** without additional features like RAG or tools
- **Learning** how Nova works under the hood

> ***✋ Most of the time, the composition of the out-of-the-box AI agents will cover your needs without dealing with base agents directly***.

### Building on Base Agents

When you create a custom agent, you can:

1. Extend a base agent with your own logic
2. Combine multiple base agents for complex workflows
3. Add custom preprocessing/postprocessing to completions

### Base Agent Features

Both completion agents provide:

- **Context management** - Automatic conversation history tracking
- **Configuration** - Flexible model and agent configuration
- **OpenAI compatibility** - Works with any OpenAI-compatible API

### Learn More

To see base agents in action, check out the samples in the repository:
- [samples/01-simple-completion/](samples/01-simple-completion/) - Basic completion example
- [samples/02-stream-completion/](samples/02-stream-completion/) - Streaming completion example
