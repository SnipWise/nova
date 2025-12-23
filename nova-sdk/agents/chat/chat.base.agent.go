package chat

import (
	"context"

	"github.com/openai/openai-go/v3"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
)

// BaseAgent wraps the shared base.Agent for chat-specific functionality
type BaseAgent struct {
	*base.Agent
}

type AgentOption func(*BaseAgent)

// NewBaseAgent creates a new ChatAgent instance using the shared base agent
func NewBaseAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption,
) (chatAgent *BaseAgent, err error) {

	// Create the shared base agent
	baseAgent, err := base.NewAgent(ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	chatAgent = &BaseAgent{
		Agent: baseAgent,
	}

	// Apply chat-specific options
	for _, option := range options {
		option(chatAgent)
	}

	return chatAgent, nil
}

func (agent *BaseAgent) Kind() (kind agents.Kind) {
	return agents.Chat
}
