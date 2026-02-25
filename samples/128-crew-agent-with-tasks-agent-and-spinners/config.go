package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig is the top-level configuration loaded from config.yaml.
type AppConfig struct {
	EngineURL        string                 `yaml:"engine_url"`
	LogLevel         string                 `yaml:"log_level"`
	ContextSizeLimit int                    `yaml:"context_size_limit"`
	Agents           map[string]AgentConfig `yaml:"agents"`
	Routing          RoutingConfig          `yaml:"routing"`
	Tools            []ToolConfig           `yaml:"tools"`
	Skills           []SkillConfig          `yaml:"skills"`
}

// SkillConfig describes a reusable skill invocable via /skill <name>.
// The Prompt field is a template: {{input}} is replaced with the user's arguments.
type SkillConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Prompt      string `yaml:"prompt"`
}

// AgentConfig describes a single agent's model and instructions.
type AgentConfig struct {
	Model        string  `yaml:"model"`
	Temperature  float64 `yaml:"temperature"`
	Instructions string  `yaml:"instructions"`
}

// RoutingConfig describes topic-based routing rules.
type RoutingConfig struct {
	Rules        []RoutingRule `yaml:"rules"`
	DefaultAgent string        `yaml:"default_agent"`
}

// RoutingRule maps a list of topics to a target agent ID.
type RoutingRule struct {
	Topics []string `yaml:"topics"`
	Agent  string   `yaml:"agent"`
}

// ToolConfig describes a tool available to the tools agent.
// Command is a shell command template with {{param}} placeholders.
type ToolConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Command     string            `yaml:"command"`
	Parameters  []ToolParamConfig `yaml:"parameters"`
}

// findToolConfig returns the ToolConfig for a given tool name, or nil if not found.
func (cfg *AppConfig) findToolConfig(name string) *ToolConfig {
	for i := range cfg.Tools {
		if cfg.Tools[i].Name == name {
			return &cfg.Tools[i]
		}
	}
	return nil
}

// findSkill returns the SkillConfig for a given skill name, or nil if not found.
func (cfg *AppConfig) findSkill(name string) *SkillConfig {
	for i := range cfg.Skills {
		if cfg.Skills[i].Name == name {
			return &cfg.Skills[i]
		}
	}
	return nil
}

// ToolParamConfig describes a single parameter for a tool.
type ToolParamConfig struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

// loadConfig reads and parses a YAML configuration file.
func loadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return &cfg, nil
}

// getAgentConfig returns the AgentConfig for a given agent key, or an error if missing.
func (cfg *AppConfig) getAgentConfig(key string) (AgentConfig, error) {
	ac, ok := cfg.Agents[key]
	if !ok {
		return AgentConfig{}, fmt.Errorf("agent %q not found in config", key)
	}
	return ac, nil
}
