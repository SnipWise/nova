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
		KeepConversationHistory: true,
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


## Installation

```bash
go get github.com/snipwise/nova@latest
```
