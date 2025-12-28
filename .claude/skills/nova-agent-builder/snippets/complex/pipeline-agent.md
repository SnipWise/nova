---
id: pipeline-agent
name: Pipeline Agent (Chained Agents)
category: complex
complexity: advanced
sample_source: 56
description: System of chained agents with transformations between each step
interactive: true
---

# Pipeline Agent (Chained Agents)

## Description

Creates a pipeline of agents where the output of each step becomes the input of the next. Ideal for complex workflows requiring multiple sequential transformations.

## Use Cases

- Document processing pipelines
- Data transformation workflows
- Content generation with review
- Multi-step analysis
- ETL-like AI processes

## ⚠️ Interactive Mode

This snippet requires specific information. Answer the following questions:

### Configuration Questions

1. **How many steps in your pipeline?**
   - Typically 2-5 steps
   - Each step = one agent

2. **What transformation at each step?**
   - Examples: extract, transform, validate, summarize, format

3. **Branching conditions?**
   - Linear (A → B → C)
   - Conditional (if X then B else C)

4. **Error handling?**
   - Stop on error
   - Skip and continue
   - Retry with fallback

5. **Parallelization?**
   - Sequential only
   - Parallel branches
   - Fan-out/fan-in

---

## Base Template

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// === PIPELINE STEP ===
type PipelineStep struct {
	Name        string
	Agent       *chat.Agent
	Transform   func(input string) string // Optional pre-processing
	Validate    func(output string) error // Optional validation
}

// === PIPELINE ===
type Pipeline struct {
	steps []PipelineStep
}

func NewPipeline() *Pipeline {
	return &Pipeline{steps: []PipelineStep{}}
}

func (p *Pipeline) AddStep(step PipelineStep) {
	p.steps = append(p.steps, step)
}

// Execute runs the pipeline
func (p *Pipeline) Execute(ctx context.Context, input string) (string, error) {
	current := input

	fmt.Println("🚀 Starting pipeline execution")
	fmt.Println(strings.Repeat("=", 50))

	for i, step := range p.steps {
		fmt.Printf("\n📍 Step %d/%d: %s\n", i+1, len(p.steps), step.Name)
		fmt.Println(strings.Repeat("-", 30))

		// Apply pre-transformation if defined
		if step.Transform != nil {
			current = step.Transform(current)
		}

		// Execute agent
		result, err := step.Agent.GenerateCompletion([]messages.Message{
			{Role: roles.User, Content: current},
		})
		if err != nil {
			return "", fmt.Errorf("step '%s' failed: %v", step.Name, err)
		}

		// Validate output if validator defined
		if step.Validate != nil {
			if err := step.Validate(result.Response); err != nil {
				return "", fmt.Errorf("step '%s' validation failed: %v", step.Name, err)
			}
		}

		fmt.Printf("✅ Output: %s\n", truncate(result.Response, 100))
		current = result.Response
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🏁 Pipeline completed")

	return current, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// === CREATE PIPELINE EXAMPLE ===
func createDocumentPipeline(ctx context.Context, engineURL string) (*Pipeline, error) {
	pipeline := NewPipeline()

	// Step 1: Extract key information
	extractor, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "extractor",
			EngineURL:               engineURL,
			KeepConversationHistory: true,
			SystemInstructions: `Extract the key information from the text:
- Main topic
- Key facts
- Important entities
- Dates and numbers

Format as a structured list.`,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.2),
		},
	)
	if err != nil {
		return nil, err
	}
	pipeline.AddStep(PipelineStep{
		Name:  "Extract",
		Agent: extractor,
	})

	// Step 2: Summarize
	summarizer, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "summarizer",
			EngineURL:               engineURL,
			KeepConversationHistory: true,
			SystemInstructions: `Create a concise summary from the extracted information.
Keep only essential points. Maximum 3 sentences.`,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.3),
		},
	)
	if err != nil {
		return nil, err
	}
	pipeline.AddStep(PipelineStep{
		Name:  "Summarize",
		Agent: summarizer,
	})

	// Step 3: Format
	formatter, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "formatter",
			EngineURL:               engineURL,
			KeepConversationHistory: true,
			SystemInstructions: `Format the summary as a professional brief:
- Title
- Executive summary
- Key points (bullet list)`,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.5),
		},
	)
	if err != nil {
		return nil, err
	}
	pipeline.AddStep(PipelineStep{
		Name:  "Format",
		Agent: formatter,
		Validate: func(output string) error {
			if len(output) < 50 {
				return fmt.Errorf("output too short")
			}
			return nil
		},
	})

	return pipeline, nil
}

// === MAIN ===
func main() {
	ctx := context.Background()

	pipeline, err := createDocumentPipeline(ctx, "http://localhost:12434/engines/llama.cpp/v1")
	if err != nil {
		fmt.Printf("Pipeline creation error: %v\n", err)
		return
	}

	// Test document
	document := `
	Apple Inc. announced today that its Q4 2024 revenue reached $95 billion, 
	exceeding analyst expectations by 5%. CEO Tim Cook attributed the growth 
	to strong iPhone 16 sales, particularly in China where sales grew 12%.
	The company also announced a $100 billion share buyback program and 
	increased its dividend by 4%. The stock rose 3% in after-hours trading.
	`

	result, err := pipeline.Execute(ctx, document)
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}

	fmt.Println("\n📄 Final Result:")
	fmt.Println(result)
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "ai/qwen2.5:1.5B-F16"

# Temperature per step (lower = more deterministic)
EXTRACTOR_TEMP: 0.2
SUMMARIZER_TEMP: 0.3
FORMATTER_TEMP: 0.5
```

## Customization

### With Conditional Branching

```go
type ConditionalPipeline struct {
    steps     []PipelineStep
    branches  map[string][]PipelineStep
    condition func(output string) string // Returns branch name
}

func (cp *ConditionalPipeline) Execute(ctx context.Context, input string) (string, error) {
    current := input
    
    for _, step := range cp.steps {
        result, _ := step.Agent.GenerateCompletion(...)
        current = result.Response
        
        // Check for branching
        if cp.condition != nil {
            branch := cp.condition(current)
            if branchSteps, ok := cp.branches[branch]; ok {
                // Execute branch steps
                for _, branchStep := range branchSteps {
                    result, _ := branchStep.Agent.GenerateCompletion(...)
                    current = result.Response
                }
            }
        }
    }
    
    return current, nil
}
```

### With Parallel Steps

```go
func (p *Pipeline) ExecuteParallel(ctx context.Context, inputs []string) ([]string, error) {
    results := make([]string, len(inputs))
    var wg sync.WaitGroup
    var mu sync.Mutex
    var firstErr error
    
    for i, input := range inputs {
        wg.Add(1)
        go func(idx int, inp string) {
            defer wg.Done()
            result, err := p.Execute(ctx, inp)
            
            mu.Lock()
            if err != nil && firstErr == nil {
                firstErr = err
            }
            results[idx] = result
            mu.Unlock()
        }(i, input)
    }
    
    wg.Wait()
    return results, firstErr
}
```

### With Retry on Failure

```go
type RetryableStep struct {
    PipelineStep
    MaxRetries int
    RetryDelay time.Duration
}

func (rs *RetryableStep) Execute(ctx context.Context, input string) (string, error) {
    var lastErr error
    
    for attempt := 0; attempt <= rs.MaxRetries; attempt++ {
        result, err := rs.Agent.GenerateCompletion(...)
        if err == nil {
            if rs.Validate == nil || rs.Validate(result.Response) == nil {
                return result.Response, nil
            }
        }
        lastErr = err
        time.Sleep(rs.RetryDelay)
    }
    
    return "", fmt.Errorf("failed after %d attempts: %v", rs.MaxRetries+1, lastErr)
}
```

## Important Notes

- Each step should have a single, clear responsibility
- Lower temperature for extraction/validation, higher for creative steps
- Add validation between steps for data quality
- Consider timeout for long pipelines
- Log intermediate results for debugging
