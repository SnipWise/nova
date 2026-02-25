package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents/crew"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

func main() {
	// --- Load configuration ---
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	os.Setenv("NOVA_LOG_LEVEL", cfg.LogLevel)

	ctx := context.Background()

	// --- Create all agents from config ---
	tasksAgent, err := createTasksAgent(ctx, cfg)
	if err != nil {
		panic(err)
	}

	toolsAgent, err := createToolsAgent(ctx, cfg)
	if err != nil {
		panic(err)
	}

	chatAgent, err := createChatAgent(ctx, cfg)
	if err != nil {
		panic(err)
	}

	orchestratorAgent, err := createOrchestratorAgent(ctx, cfg)
	if err != nil {
		panic(err)
	}

	compressorAgent, err := createCompressorAgent(ctx, cfg)
	if err != nil {
		panic(err)
	}

	crewSpinner := spinner.NewWithColor("").
		SetFrameColor(spinner.ColorCyan).
		SetFrames(spinner.FramesDots).
		SetSuffix("Generating response...").
		SetSuffixColor(spinner.ColorBold + spinner.ColorBrightCyan)

	// --- Assemble the Crew Agent ---
	crewAgent, err := crew.NewAgent(
		ctx,
		crew.WithAgentCrew(chatAgent, cfg.Routing.DefaultAgent),
		crew.WithOrchestratorAgent(orchestratorAgent),
		crew.WithTasksAgent(tasksAgent),
		crew.WithToolsAgent(toolsAgent),
		crew.WithExecuteFn(makeExecuteFunction(cfg)),
		crew.WithConfirmationPromptFn(confirmationPromptFunction),
		crew.WithCompressorAgentAndContextSize(compressorAgent, cfg.ContextSizeLimit),
		crew.BeforeCompletion(func(a *crew.CrewAgent) {
			crewSpinner.Start()
		}),
		crew.AfterCompletion(func(a *crew.CrewAgent) {
			crewSpinner.Success(fmt.Sprintf("Completed! (agent: %s)", a.GetSelectedAgentId()))
		}),
	)
	if err != nil {
		panic(err)
	}

	// --- Banner ---
	fmt.Println("🤖 Crew Agent with Tasks Agent & Spinners (configurable)")
	fmt.Println(strings.Repeat("─", 55))
	fmt.Println("Spinners identify each agent:")
	fmt.Printf("  ⣾ %s%sAnalyzing your plan...%s\n", spinner.ColorBold, spinner.ColorBrightCyan, spinner.ColorReset)
	fmt.Printf("  ⣾ %s%sSelecting the agent...%s\n", spinner.ColorBold, spinner.ColorBrightCyan, spinner.ColorReset)
	fmt.Printf("  ⣾ %s%sGenerating response...%s\n", spinner.ColorBold, spinner.ColorBrightCyan, spinner.ColorReset)
	fmt.Printf("  ⣾ %s%sExecuting the tool...%s\n", spinner.ColorBold, spinner.ColorBrightCyan, spinner.ColorReset)
	fmt.Printf("  ⣾ %s%sCompressing context...%s (limit: %d)\n", spinner.ColorBold, spinner.ColorBrightCyan, spinner.ColorReset, cfg.ContextSizeLimit)
	fmt.Println(strings.Repeat("─", 55))
	fmt.Println("Commands:")
	fmt.Println("  /new            → Reset all agents memory and start fresh")
	fmt.Println("  /pack           → Force context packing (compress history)")
	fmt.Println("  /skill <name>   → Run a skill (see below)")
	fmt.Println("  /bye            → Exit the program")
	fmt.Println("  /help           → Show available commands")
	if len(cfg.Skills) > 0 {
		fmt.Println()
		fmt.Println("Skills:")
		for _, sk := range cfg.Skills {
			fmt.Printf("  %-10s → %s\n", sk.Name, sk.Description)
		}
	}
	fmt.Println()

	// --- REPL loop ---
	for {
		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🧑 You: ")
		question, err := input.RunWithEdit()
		if err != nil {
			display.Errorf("Error reading input: %v", err)
			continue
		}

		if strings.HasPrefix(question, "/bye") {
			display.Infof("👋 Goodbye!")
			break
		}

		if strings.HasPrefix(question, "/new") {
			crewAgent.ResetMessages()
			display.Infof("🧹 All agents memory has been reset. Starting fresh!")
			display.NewLine()
			continue
		}

		if strings.HasPrefix(question, "/pack") {
			newSize, packErr := crewAgent.CompressChatAgentContext()
			if packErr != nil {
				display.Errorf("❌ Context packing failed: %v", packErr)
			} else {
				display.Infof("🗜️  Context packed! New size: %d characters", newSize)
			}
			display.NewLine()
			continue
		}

		if strings.HasPrefix(question, "/skill") {
			parts := strings.SplitN(strings.TrimSpace(question), " ", 3)
			if len(parts) < 2 {
				display.Errorf("Usage: /skill <name> <input>")
				display.NewLine()
				fmt.Println("Available skills:")
				for _, sk := range cfg.Skills {
					fmt.Printf("  %-10s → %s\n", sk.Name, sk.Description)
				}
				display.NewLine()
				continue
			}
			skill := cfg.findSkill(parts[1])
			if skill == nil {
				display.Errorf("Unknown skill: %s", parts[1])
				display.NewLine()
				fmt.Println("Available skills:")
				for _, sk := range cfg.Skills {
					fmt.Printf("  %-10s → %s\n", sk.Name, sk.Description)
				}
				display.NewLine()
				continue
			}
			skillInput := ""
			if len(parts) == 3 {
				skillInput = parts[2]
			}
			question = strings.ReplaceAll(skill.Prompt, "{{input}}", skillInput)
			display.Infof("Skill: %s", skill.Name)
			display.NewLine()
			// Fall through to normal completion with the enriched prompt
		}

		if strings.HasPrefix(question, "/help") {
			display.NewLine()
			fmt.Println("Available commands:")
			fmt.Println("  /new            → Reset all agents memory and start fresh")
			fmt.Println("  /pack           → Force context packing (compress history)")
			fmt.Println("  /skill <name>   → Run a skill (see below)")
			fmt.Println("  /bye            → Exit the program")
			fmt.Println("  /help           → Show this help message")
			display.NewLine()
			if len(cfg.Skills) > 0 {
				fmt.Println("Available skills:")
				for _, sk := range cfg.Skills {
					fmt.Printf("  %-10s → %s\n", sk.Name, sk.Description)
				}
				display.NewLine()
			}
			continue
		}

		if question == "" {
			continue
		}

		display.NewLine()

		result, err := crewAgent.StreamCompletion(question, func(chunk string, finishReason string) error {
			// Stop the crew spinner on first chunk (the LLM is now streaming)
			if crewSpinner.IsRunning() {
				crewSpinner.Stop()
			}
			if chunk != "" {
				display.MarkdownChunk(markdownParser, chunk)
			}
			if finishReason == "stop" {
				markdownParser.Flush()
				markdownParser.Reset()
				display.NewLine()
			}
			return nil
		})

		if err != nil {
			// Make sure spinners are stopped on error
			if crewSpinner.IsRunning() {
				crewSpinner.Error("Error occurred")
			}
			display.Errorf("❌ Error: %v", err)
			continue
		}

		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", crewAgent.GetContextSize()))
		display.Separator()
	}
}
