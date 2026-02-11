package agents

import (
	"context"
	"errors"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// normalizeModelName removes common prefixes and suffixes to enable flexible model name matching
// - Removes "docker.io/" prefix if present
// - Removes ":latest" suffix if present (but keeps other tags like ":0.5B-F16")
// Examples:
//   - "docker.io/ai/mxbai-embed-large:latest" -> "ai/mxbai-embed-large"
//   - "ai/mxbai-embed-large" -> "ai/mxbai-embed-large"
//   - "docker.io/ai/qwen2.5:0.5B-F16" -> "ai/qwen2.5:0.5B-F16" (keeps the tag)
func normalizeModelName(modelName string) string {
	// Remove "docker.io/" prefix
	normalized := strings.TrimPrefix(modelName, "docker.io/")

	// Remove ":latest" suffix only (not other tags)
	normalized = strings.TrimSuffix(normalized, ":latest")

	return normalized
}

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

	// Check if the model is available on the specified engine URL
	// Uses normalizeModelName to handle variations like:
	// - "ai/mxbai-embed-large" matching "docker.io/ai/mxbai-embed-large:latest"
	// - "docker.io/ai/qwen2.5:0.5B-F16" matching "ai/qwen2.5:0.5B-F16"
	modelsList := client.Models.ListAutoPaging(ctx)
	modelFound := false
	normalizedSearchName := normalizeModelName(modelConfig.Name)

	for modelsList.Next() {
		m := modelsList.Current()
		normalizedModelID := normalizeModelName(m.ID)

		log.Debug("🔎 Comparing: '%s' (from '%s') with '%s' (from '%s')",
			normalizedModelID, m.ID, normalizedSearchName, modelConfig.Name)

		if normalizedModelID == normalizedSearchName {
			log.Debug("✅ Model matched: '%s' matches '%s'", m.ID, modelConfig.Name)
			modelFound = true
			break
		}
	}

	if err := modelsList.Err(); err != nil {
		log.Error("Error listing models: %v", err)
		return openai.Client{}, nil, err
	}

	if !modelFound {
		log.Error("Model not available: %s (normalized: %s)", modelConfig.Name, normalizedSearchName)
		return openai.Client{}, nil, errors.New("model not available on the specified engine URL")
	}

	log.Info("✅ Model %s is available on %s", modelConfig.Name, agentConfig.EngineURL)

	// _, err = client.Models.Get(ctx, modelConfig.Name)

	// if err != nil {
	// 	log.Error("Model not available:", err)
	// 	return openai.Client{}, nil, errors.New("model not available on the specified engine URL")
	// }

	// log.Info("Model %s is available on %s", modelConfig.Name, agentConfig.EngineURL)

	return client, log, nil

}
