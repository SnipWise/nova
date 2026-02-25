package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

// makeExecuteFunction returns a tool execution function that resolves
// shell commands from the YAML config. Each tool's "command" field is
// a template with {{param}} placeholders that get substituted with the
// actual arguments provided by the LLM.
//
// Adding a new tool is as simple as adding an entry in config.yaml —
// no Go code changes required.
func makeExecuteFunction(cfg *AppConfig) func(string, string) (string, error) {
	return func(functionName string, arguments string) (string, error) {
		toolSpinner := spinner.NewWithColor("").
			SetFrameColor(spinner.ColorCyan).
			SetFrames(spinner.FramesDots).
			SetSuffix(fmt.Sprintf("Executing %s...", functionName)).
			SetSuffixColor(spinner.ColorBold + spinner.ColorBrightCyan)

		toolSpinner.Start()

		// Find the tool config
		tc := cfg.findToolConfig(functionName)
		if tc == nil {
			toolSpinner.Error(fmt.Sprintf("Unknown tool: %s", functionName))
			return `{"error": "Unknown tool"}`, fmt.Errorf("tool %q not found in config", functionName)
		}

		// Parse arguments JSON into a map
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			toolSpinner.Error(fmt.Sprintf("Invalid arguments for %s", functionName))
			return `{"error": "Invalid arguments"}`, nil
		}

		// Build the shell command by substituting {{param}} placeholders
		cmd := tc.Command
		for key, val := range args {
			cmd = strings.ReplaceAll(cmd, "{{"+key+"}}", fmt.Sprintf("%v", val))
		}

		// Execute the shell command
		output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		result := strings.TrimSpace(string(output))

		if err != nil {
			toolSpinner.Error(fmt.Sprintf("%s failed", functionName))
			return fmt.Sprintf(`{"error": %q, "output": %q}`, err.Error(), result), nil
		}

		toolSpinner.Success(fmt.Sprintf("%s completed", functionName))
		return fmt.Sprintf(`{"output": %q}`, result), nil
	}
}

// confirmationPromptFunction asks the user to confirm a tool call before execution.
func confirmationPromptFunction(functionName string, arguments string) tools.ConfirmationResponse {
	display.Colorf(display.ColorGreen, "🟢 Detected function: %s with arguments: %s\n", functionName, arguments)
	choice := prompt.HumanConfirmation(fmt.Sprintf("Execute %s with %v?", functionName, arguments))
	return choice
}
