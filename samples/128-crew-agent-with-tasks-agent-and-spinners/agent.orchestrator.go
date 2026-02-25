package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

// createOrchestratorAgent creates the Orchestrator Agent for topic detection and routing.
// It reads system instructions from the .md file and routing rules from config.yaml.
// The routing config is passed via WithRoutingConfig so the crew agent can
// automatically route topics to the correct chat agent (no manual matchAgentFunction needed).
func createOrchestratorAgent(ctx context.Context, cfg *AppConfig) (*orchestrator.Agent, error) {
	ac, err := cfg.getAgentConfig("orchestrator")
	if err != nil {
		return nil, err
	}

	routingConfig := buildRoutingConfig(cfg)

	orchestratorSpinner := spinner.NewWithColor("").
		SetFrameColor(spinner.ColorCyan).
		SetFrames(spinner.FramesDots).
		SetSuffix("Selecting the agent...").
		SetSuffixColor(spinner.ColorBold + spinner.ColorBrightCyan)

	return orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:               "orchestrator-agent",
			EngineURL:          cfg.EngineURL,
			SystemInstructions: ac.Instructions,
		},
		models.Config{
			Name:        ac.Model,
			Temperature: models.Float64(ac.Temperature),
		},
		orchestrator.WithRoutingConfig(routingConfig),
		orchestrator.BeforeCompletion(func(a *orchestrator.Agent) {
			orchestratorSpinner.Start()
		}),
		orchestrator.AfterCompletion(func(a *orchestrator.Agent) {
			orchestratorSpinner.Success("Agent selected!")
		}),
	)
}

// buildRoutingConfig converts the YAML routing config to the SDK's AgentRoutingConfig.
func buildRoutingConfig(cfg *AppConfig) orchestrator.AgentRoutingConfig {
	var routes []struct {
		Topics []string `json:"topics"`
		Agent  string   `json:"agent"`
	}

	for _, rule := range cfg.Routing.Rules {
		routes = append(routes, struct {
			Topics []string `json:"topics"`
			Agent  string   `json:"agent"`
		}{
			Topics: rule.Topics,
			Agent:  rule.Agent,
		})
	}

	return orchestrator.AgentRoutingConfig{
		Routing:      routes,
		DefaultAgent: cfg.Routing.DefaultAgent,
	}
}
