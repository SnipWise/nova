package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)


// Expected results for validation
var expectedResults = []string{
	`{"result": 42}`,
	`{"message": "👋 Hello, Bob!🙂"}`,
	`{"message": "👋 Hello, Sam!🙂"}`,
	`{"result": 42}`,
	`{"message": "👋 Hello, Alice!🙂"}`,
}

// TestReport stores the results for each configuration
type TestReport struct {
	ModelName       string
	TotalCalls      int
	SuccessfulCalls int
	FailedCalls     int
	MissingCalls    int
	ExtraCalls      int
	TimedOut        bool
	HasError        bool
	ErrorMessage    string
	Duration        time.Duration
	Details         []string
}

func main() {
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}
	ctx := context.Background()
	engineURL := "http://localhost:12434/engines/llama.cpp/v1"

	configList := []models.Config{
		{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
		{
			Name:              "huggingface.co/unsloth/functiongemma-270m-it-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},		
		{
			Name:              "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
		{
			Name:              "ai/qwen2.5:1.5B-F16",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},		
	}

	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL: engineURL,
			SystemInstructions: `
				You are Bob, a helpful AI assistant.
				You have access to some functions that you can call to help the user.
			`,
		},
		models.Config{
			Name: "ai/qwen2.5:1.5B-F16",
		},

		tools.WithTools(GetToolsIndex()),
	)

	if err != nil {
		panic(err)
	}

	// Demo: Test with multiple model configurations using SetModelConfig
	display.Colorf(display.ColorYellow, "\n========================================\n")
	display.Colorf(display.ColorYellow, "Testing with different model configs\n")
	display.Colorf(display.ColorYellow, "========================================\n\n")

	var reports []TestReport

	for i, modelConfig := range configList {
		display.Colorf(display.ColorCyan, "\n--- Test %d: %s ---\n", i+1, modelConfig.Name)

		// Initialize report for this config
		report := TestReport{
			ModelName: modelConfig.Name,
			Details:   []string{},
		}

		// Update the model configuration using the setter
		agent.SetModelConfig(modelConfig)

		// Verify the config was updated
		currentModelConfig := agent.GetModelConfig()
		display.KeyValue("Current Model", currentModelConfig.Name)
		display.KeyValue("Temperature", fmt.Sprintf("%v", *currentModelConfig.Temperature))

		// Reset messages for each test
		agent.ResetMessages()

		messages := []messages.Message{
			{
				Content: `
				Make the sum of 40 and 2,
				then say hello to Bob and to Sam,
				make the sum of 5 and 37
				Say hello to Alice
				`,
				Role: roles.User,
			},
		}

		// Execute test with timeout
		type testResult struct {
			result *tools.ToolCallResult
			err    error
		}
		resultChan := make(chan testResult, 1)

		startTime := time.Now()

		// Run the test in a goroutine
		go func() {
			result, err := agent.DetectToolCallsLoop(
				messages,
				executeFunction,
			)
			resultChan <- testResult{result: result, err: err}
		}()

		// Wait for result or timeout (60 seconds)
		var result *tools.ToolCallResult
		var err error

		select {
		case res := <-resultChan:
			result = res.result
			err = res.err
			report.Duration = time.Since(startTime)

		case <-time.After(60 * time.Second):
			report.TimedOut = true
			report.Duration = time.Since(startTime)
			display.Colorf(display.ColorRed, "⏱️  Timeout after 60 seconds for %s\n", modelConfig.Name)
			report.Details = append(report.Details, "❌ Timeout: Model took too long to respond (>60s)")
			reports = append(reports, report)
			continue
		}

		if err != nil {
			report.HasError = true
			report.ErrorMessage = err.Error()
			display.Colorf(display.ColorRed, "❌ Error with %s: %v\n", modelConfig.Name, err)
			report.Details = append(report.Details, fmt.Sprintf("❌ Error: %v", err))
			reports = append(reports, report)
			continue
		}

		display.KeyValue("Finish Reason", result.FinishReason)

		// Validate results
		report.TotalCalls = len(result.Results)
		expectedCount := len(expectedResults)

		// Check each result
		for idx, actualResult := range result.Results {
			display.KeyValue("Result for tool", actualResult)

			if idx < expectedCount {
				// Normalize whitespace for comparison
				normalizedActual := strings.TrimSpace(actualResult)
				normalizedExpected := strings.TrimSpace(expectedResults[idx])

				if normalizedActual == normalizedExpected {
					report.SuccessfulCalls++
					report.Details = append(report.Details, fmt.Sprintf("✅ Call %d: %s", idx+1, normalizedActual))
					display.Colorf(display.ColorGreen, "  ✅ Match expected result\n")
				} else {
					report.FailedCalls++
					report.Details = append(report.Details, fmt.Sprintf("❌ Call %d: Expected %s, Got %s", idx+1, normalizedExpected, normalizedActual))
					display.Colorf(display.ColorRed, "  ❌ Expected: %s\n", normalizedExpected)
				}
			} else {
				report.ExtraCalls++
				report.Details = append(report.Details, fmt.Sprintf("⚠️  Extra call %d: %s", idx+1, actualResult))
				display.Colorf(display.ColorYellow, "  ⚠️  Extra unexpected call\n")
			}
		}

		// Check for missing calls
		if report.TotalCalls < expectedCount {
			report.MissingCalls = expectedCount - report.TotalCalls
			for idx := report.TotalCalls; idx < expectedCount; idx++ {
				report.Details = append(report.Details, fmt.Sprintf("⚠️  Missing call %d: %s", idx+1, expectedResults[idx]))
			}
			display.Colorf(display.ColorYellow, "  ⚠️  Missing %d expected call(s)\n", report.MissingCalls)
		}

		display.KeyValue("Assistant Message", result.LastAssistantMessage)
		display.Colorf(display.ColorCyan, "\n--- End Test %d ---\n", i+1)

		reports = append(reports, report)
	}

	// Print final report
	printFinalReport(reports)

}

func printFinalReport(reports []TestReport) {
	display.Colorf(display.ColorYellow, "\n\n========================================\n")
	display.Colorf(display.ColorYellow, "FINAL REPORT\n")
	display.Colorf(display.ColorYellow, "========================================\n\n")

	for i, report := range reports {
		display.Colorf(display.ColorCyan, "Report %d: %s\n", i+1, report.ModelName)
		display.Colorf(display.ColorCyan, "─────────────────────────────────────────\n")

		// Display duration
		display.KeyValue("Duration", fmt.Sprintf("%.2fs", report.Duration.Seconds()))

		// Display status
		if report.TimedOut {
			display.Colorf(display.ColorRed, "Status: ⏱️  TIMEOUT\n")
		} else if report.HasError {
			display.Colorf(display.ColorRed, "Status: ❌ ERROR\n")
			display.KeyValue("Error Message", report.ErrorMessage)
		}

		display.KeyValue("Total Calls", fmt.Sprintf("%d", report.TotalCalls))
		display.KeyValue("Successful", fmt.Sprintf("%d", report.SuccessfulCalls))
		display.KeyValue("Failed", fmt.Sprintf("%d", report.FailedCalls))
		display.KeyValue("Missing", fmt.Sprintf("%d", report.MissingCalls))
		display.KeyValue("Extra", fmt.Sprintf("%d", report.ExtraCalls))

		// Calculate success rate
		expectedTotal := len(expectedResults)
		successRate := 0.0
		if expectedTotal > 0 && !report.TimedOut && !report.HasError {
			successRate = (float64(report.SuccessfulCalls) / float64(expectedTotal)) * 100
		}

		color := display.ColorGreen
		if successRate < 100 || report.TimedOut || report.HasError {
			color = display.ColorRed
		}

		if report.TimedOut {
			display.Colorf(color, "Success Rate: 0.0%% (TIMEOUT)\n")
		} else if report.HasError {
			display.Colorf(color, "Success Rate: 0.0%% (ERROR)\n")
		} else {
			display.Colorf(color, "Success Rate: %.1f%%\n", successRate)
		}

		// Print details
		display.Colorf(display.ColorWhite, "\nDetails:\n")
		for _, detail := range report.Details {
			fmt.Println("  " + detail)
		}

		display.Colorf(display.ColorCyan, "\n")
	}

	// Comparison Table
	display.Colorf(display.ColorYellow, "\n========================================\n")
	display.Colorf(display.ColorYellow, "MODELS COMPARISON TABLE\n")
	display.Colorf(display.ColorYellow, "========================================\n\n")

	// Print table header
	fmt.Printf("%-50s | %10s | %8s | %6s | %7s | %6s | %10s | %10s\n",
		"Model Name", "Duration", "Success", "Failed", "Missing", "Extra", "Status", "Rate")
	fmt.Println(strings.Repeat("-", 140))

	// Print each model's results
	expectedTotal := len(expectedResults)
	for _, report := range reports {
		modelName := report.ModelName
		if len(modelName) > 48 {
			modelName = modelName[:45] + "..."
		}

		duration := fmt.Sprintf("%.2fs", report.Duration.Seconds())
		success := fmt.Sprintf("%d", report.SuccessfulCalls)
		failed := fmt.Sprintf("%d", report.FailedCalls)
		missing := fmt.Sprintf("%d", report.MissingCalls)
		extra := fmt.Sprintf("%d", report.ExtraCalls)

		var status string
		var rate string

		if report.TimedOut {
			status = "[TIMEOUT]"
			rate = "0.0%"
		} else if report.HasError {
			status = "[ERROR]"
			rate = "0.0%"
		} else {
			successRate := 0.0
			if expectedTotal > 0 {
				successRate = (float64(report.SuccessfulCalls) / float64(expectedTotal)) * 100
			}
			if successRate == 100.0 {
				status = "[PASS]"
			} else {
				status = "[FAIL]"
			}
			rate = fmt.Sprintf("%.1f%%", successRate)
		}

		fmt.Printf("%-50s | %10s | %8s | %6s | %7s | %6s | %10s | %10s\n",
			modelName, duration, success, failed, missing, extra, status, rate)
	}

	fmt.Println(strings.Repeat("-", 140))

	// Summary
	display.Colorf(display.ColorYellow, "\n========================================\n")
	display.Colorf(display.ColorYellow, "SUMMARY\n")
	display.Colorf(display.ColorYellow, "========================================\n")

	totalConfigs := len(reports)
	perfectConfigs := 0
	timedOutConfigs := 0
	errorConfigs := 0

	for _, report := range reports {
		if report.TimedOut {
			timedOutConfigs++
		} else if report.HasError {
			errorConfigs++
		} else if report.SuccessfulCalls == len(expectedResults) && report.FailedCalls == 0 && report.MissingCalls == 0 && report.ExtraCalls == 0 {
			perfectConfigs++
		}
	}

	display.KeyValue("Total Configurations Tested", fmt.Sprintf("%d", totalConfigs))
	display.KeyValue("Perfect Results", fmt.Sprintf("%d", perfectConfigs))
	display.KeyValue("Timeouts", fmt.Sprintf("%d", timedOutConfigs))
	display.KeyValue("Errors", fmt.Sprintf("%d", errorConfigs))
	display.KeyValue("Failed or Incomplete", fmt.Sprintf("%d", totalConfigs-perfectConfigs-timedOutConfigs-errorConfigs))

	if perfectConfigs == totalConfigs {
		display.Colorf(display.ColorGreen, "\n✅ ALL CONFIGURATIONS PASSED!\n\n")
	} else if timedOutConfigs > 0 {
		display.Colorf(display.ColorRed, "\n⏱️  Some configurations timed out\n\n")
	} else if errorConfigs > 0 {
		display.Colorf(display.ColorRed, "\n❌ Some configurations had errors\n\n")
	} else {
		display.Colorf(display.ColorRed, "\n❌ Some configurations failed or incomplete\n\n")
	}
}

func GetToolsIndex() []*tools.Tool {

	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {

	display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s!🙂", args.Name)
		return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	case "calculate_sum":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for calculate_sum"}`, nil
		}
		sum := args.A + args.B
		return fmt.Sprintf(`{"result": %g}`, sum), nil

	case "say_exit":

		// NOTE: Returning a message and an ExitToolCallsLoopError to stop further processing
		return fmt.Sprintf(`{"message": "%s"}`, "❌ EXIT"), errors.New("exit_loop")

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
