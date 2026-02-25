package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

// createTasksAgent creates the Tasks Agent with a planner spinner.
func createTasksAgent(ctx context.Context, cfg *AppConfig) (*tasks.Agent, error) {
	ac, err := cfg.getAgentConfig("tasks_planner")
	if err != nil {
		return nil, err
	}

	plannerSpinner := spinner.NewWithColor("").
		SetFrameColor(spinner.ColorCyan).
		SetFrames(spinner.FramesDots).
		SetSuffix("Analyzing your plan...").
		SetSuffixColor(spinner.ColorBold + spinner.ColorBrightCyan)

	return tasks.NewAgent(
		ctx,
		agents.Config{
			Name:               "project-planner",
			EngineURL:          cfg.EngineURL,
			SystemInstructions: ac.Instructions,
		},
		models.Config{
			Name:        ac.Model,
			Temperature: models.Float64(ac.Temperature),
		},
		tasks.BeforeCompletion(func(a *tasks.Agent) {
			plannerSpinner.Start()
		}),
		tasks.AfterCompletion(func(a *tasks.Agent) {
			plannerSpinner.Success("Plan identified!")
		}),
	)
}
