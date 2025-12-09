package rag

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/toolbox/logger"
)

type BaseAgent struct {
	ctx             context.Context
	config          agents.AgentConfig
	EmbeddingParams openai.EmbeddingNewParams
	openaiClient    openai.Client
	log             logger.Logger
}

type AgentOption func(*BaseAgent)

func NewBaseAgent(
	ctx context.Context,
	agentConfig agents.AgentConfig,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption,
) (ragAgent *BaseAgent, err error) {
	// export SNIP_LOG_LEVEL=debug  # Shows all logs
	// export SNIP_LOG_LEVEL=info   # Shows info, warn, error
	// export SNIP_LOG_LEVEL=warn   # Shows warn, error only
	// export SNIP_LOG_LEVEL=error  # Shows errors only
	// export SNIP_LOG_LEVEL=none   # No logging (default)

	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	client := openai.NewClient(
		option.WithBaseURL(agentConfig.EngineURL),
		option.WithAPIKey("I💙DockerModelRunner"),
	)

	_, err = client.Models.Get(ctx, modelConfig.Model)
	if err != nil {
		log.Error("Model not available:", err)
		return nil, errors.New("model not available on the specified engine URL")
	}
	log.Info("Model %s is available on %s", modelConfig.Model, agentConfig.EngineURL)

}
