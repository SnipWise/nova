# Sample 123 - Tasks Agent with Spinner Feedback

This sample builds on [sample 122](../122-tasks-agent/) and adds a **spinner with random thinking messages** to provide visual feedback while the LLM generates the execution plan.

## What's New Compared to Sample 122

Sample 122 uses simple `fmt.Println` in the lifecycle hooks. This sample replaces them with the N.O.V.A. `spinner` package to display an animated Braille spinner with rotating status messages, similar to what you see in modern CLI tools while they "think".

### Terminal Output During Generation

```
🧠 ⠹ Detecting dependencies...
🧠 ⠧ Ordering tasks by priority...
🧠 ⠏ Building the execution plan...
✓ Plan identification completed!
```

The spinner animates the Braille frames at 100ms intervals while the suffix text changes every 2 seconds with a random thinking message.

## How It Works

### The Spinner + Goroutine Pattern

The implementation uses three concurrent elements:

```go
// 1. Thinking messages pool
thinkingMessages := []string{
    "Analyzing the request...",
    "Identifying tasks...",
    "Detecting dependencies...",
    "Ordering tasks by priority...",
    "Classifying tool vs completion tasks...",
    "Evaluating task complexity...",
    "Building the execution plan...",
    "Checking prerequisites...",
    "Almost there...",
}

// 2. Spinner (prefix = brain emoji, suffix = rotating messages)
s := spinner.New("🧠")

// 3. Channel to signal the message rotation goroutine to stop
stopThinking := make(chan bool)
```

### Lifecycle Hooks

The spinner is wired into the Tasks Agent via `BeforeCompletion` and `AfterCompletion` hooks:

**BeforeCompletion** — starts the spinner and launches a goroutine that rotates thinking messages:

```go
tasks.BeforeCompletion(func(a *tasks.Agent) {
    s.Start()
    go func() {
        for {
            select {
            case <-stopThinking:
                return
            default:
                msg := thinkingMessages[rand.Intn(len(thinkingMessages))]
                s.SetSuffix(msg)
                time.Sleep(2 * time.Second)
            }
        }
    }()
}),
```

**AfterCompletion** — stops the message rotation, then stops the spinner with a success message:

```go
tasks.AfterCompletion(func(a *tasks.Agent) {
    stopThinking <- true
    s.Success("Plan identification completed!")
}),
```

### Concurrency Model

```
BeforeCompletion                          AfterCompletion
      │                                        │
      ├─► spinner.Start()                      ├─► stopThinking <- true
      │     └─► internal goroutine             │     └─► message goroutine exits
      │           (Braille frames @ 100ms)     │
      │                                        └─► spinner.Success(...)
      └─► message goroutine                          └─► spinner.Stop()
            (random suffix @ 2s)                           └─► internal goroutine exits
```

Three goroutines cooperate:
1. **Spinner goroutine** (internal to the SDK) — animates Braille characters at 100ms
2. **Message goroutine** (launched in BeforeCompletion) — updates the suffix every 2 seconds
3. **Main goroutine** — runs `IdentifyPlanFromText`, which blocks until the LLM responds

`SetSuffix` is thread-safe (protected by `sync.RWMutex` in the SDK), so the message goroutine can safely update the text while the spinner reads it for display.

### Closure-Based State Sharing

The spinner and the `stopThinking` channel are declared in `main()` scope and captured by the hook closures. This avoids adding metadata storage to the Tasks Agent struct — the Go closure mechanism naturally shares state between the two hooks.

## The Task Struct

Same as sample 122. See the [sample 122 README](../122-tasks-agent/README.md) for full documentation of:
- The `Task` struct fields (`ID`, `Description`, `Responsible`, `ToolName`, `Arguments`, `DependsOn`, `Complexity`)
- The three responsible types (`"tool"`, `"completion"`, `"developer"`)
- Complexity levels (`"simple"`, `"moderate"`, `"complex"`)
- Dependency-based ordering rules and the few-shot example
- The expected output and dependency graph

## The N.O.V.A. Spinner Package

The `spinner` package (`nova-sdk/ui/spinner/`) provides two spinner types:

| Type | Constructor | Features |
|---|---|---|
| `Spinner` | `spinner.New(prefix)` | Basic spinner with prefix/suffix |
| `ColorSpinner` | `spinner.NewWithColor(prefix)` | Spinner with ANSI color support |

Available frame sets:

| Variable | Style |
|---|---|
| `FramesBraille` | `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏` (default) |
| `FramesDots` | `⣾ ⣽ ⣻ ⢿ ⡿ ⣟ ⣯ ⣷` |
| `FramesPulsingStar` | `✦ ✶ ✷ ✸ ✹ ✸ ✷ ✶` |
| `FramesArrows` | `← ↖ ↑ ↗ → ↘ ↓ ↙` |
| `FramesCircle` | `◐ ◓ ◑ ◒` |
| `FramesASCII` | `\| / - \` |
| `FramesProgressive` | `. .. ... .... .....` |

## Running the Sample

```bash
go run samples/123-tasks-agent/main.go
```

Requires a local LLM engine running at `http://localhost:12434/engines/llama.cpp/v1`.
