package agents

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

func InitializeConnection(ctx context.Context, agentConfig Config, modelConfig models.Config) (client openai.Client, log logger.Logger, err error) {
	// export NOVA_LOG_LEVEL=debug  # Shows all logs
	// export NOVA_LOG_LEVEL=info   # Shows info, warn, error
	// export NOVA_LOG_LEVEL=warn   # Shows warn, error only
	// export NOVA_LOG_LEVEL=error  # Shows errors only
	// export NOVA_LOG_LEVEL=none   # No logging (default)

	// Create logger from environment variable
	log = logger.GetLoggerFromEnv()

	client = openai.NewClient(
		option.WithBaseURL(agentConfig.EngineURL),
		option.WithAPIKey(agentConfig.APIKey),
	)

	_, err = client.Models.Get(ctx, modelConfig.Name)
	if err != nil {
		log.Error("Model not available:", err)
		return openai.Client{}, nil, errors.New("model not available on the specified engine URL")
	}
	log.Info("Model %s is available on %s", modelConfig.Name, agentConfig.EngineURL)

	return client, log, nil

}
