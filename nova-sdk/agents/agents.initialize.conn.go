package agents

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

func InitializeConnection(ctx context.Context, engineURL, model string) (client openai.Client, log logger.Logger, err error) {
	// export SNIP_LOG_LEVEL=debug  # Shows all logs
	// export SNIP_LOG_LEVEL=info   # Shows info, warn, error
	// export SNIP_LOG_LEVEL=warn   # Shows warn, error only
	// export SNIP_LOG_LEVEL=error  # Shows errors only
	// export SNIP_LOG_LEVEL=none   # No logging (default)

	// Create logger from environment variable
	log = logger.GetLoggerFromEnv()

	client = openai.NewClient(
		option.WithBaseURL(engineURL),
		option.WithAPIKey("I💙DockerModelRunner"),
	)

	_, err = client.Models.Get(ctx, model)
	if err != nil {
		log.Error("Model not available:", err)
		return openai.Client{}, nil, errors.New("model not available on the specified engine URL")
	}
	log.Info("Model %s is available on %s", model, engineURL)

	return client, log, nil

}
