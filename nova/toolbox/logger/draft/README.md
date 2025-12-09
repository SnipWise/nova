# Logger Package

A flexible logging system for the snip-sdk that allows you to control log output through environment variables or programmatically.

## Features

- **Multiple log levels**: DEBUG, INFO, WARN, ERROR, NONE
- **Environment variable control**: Set `SNIP_LOG_LEVEL` to control logging globally
- **NoOp logger**: Default logger that produces no output (zero overhead)
- **Console logger**: Logs to stdout/stderr with formatted output
- **Custom prefix support**: Add agent name or component prefix to logs
- **Interface-based**: Easy to implement custom loggers

## Usage

### Using Environment Variable

Set the `SNIP_LOG_LEVEL` environment variable:

```bash
export SNIP_LOG_LEVEL=debug  # Shows all logs
export SNIP_LOG_LEVEL=info   # Shows info, warn, error
export SNIP_LOG_LEVEL=warn   # Shows warn, error only
export SNIP_LOG_LEVEL=error  # Shows errors only
export SNIP_LOG_LEVEL=none   # No logging (default)
```

Then use the logger in your code:

```go
import "github.com/snipwise/snip-sdk/snip/toolbox/logger"

// Create logger from environment variable
log := logger.GetLoggerFromEnv()

log.Debug("Detailed debugging info")
log.Info("General information")
log.Warn("Warning message")
log.Error("Error occurred: %v", err)
```

### Using with Agent Prefix

```go
// Create logger with agent name prefix
log := logger.GetLoggerFromEnvWithPrefix("MyAgent")

log.Info("This will show: [MyAgent] [INFO] This will show")
```

### Programmatic Configuration

```go
// Create a console logger with specific level
log := logger.NewConsoleLogger(logger.LevelDebug)

// Create with prefix
log := logger.NewConsoleLoggerWithPrefix(logger.LevelInfo, "ChatAgent")

// Change level at runtime
log.SetLevel(logger.LevelWarn)

// Check current level
currentLevel := log.GetLevel()
```



## Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `DEBUG` | Detailed debugging information | Development, troubleshooting |
| `INFO` | General informational messages | Normal operations, confirmations |
| `WARN` | Warning messages | Potential issues, deprecated usage |
| `ERROR` | Error messages | Errors and failures |
| `NONE` | No logging | Production (default) |

## Integration with Agents

Agents can use the logger system:

```go
type ChatAgent struct {
    logger logger.Logger
    // ... other fields
}

// In NewChatAgent
agent := &ChatAgent{
    logger: logger.GetLoggerFromEnvWithPrefix(agentConfig.Name),
    // ... other initialization
}

// Usage in methods
agent.logger.Info("Model %s is available at %s", modelID, engineURL)
agent.logger.Error("Failed to connect: %v", err)
```

## Custom Logger Implementation

Implement the `Logger` interface for custom logging:

```go
type CustomLogger struct {
    // Your fields
}

func (c *CustomLogger) Debug(msg string, args ...interface{}) {
    // Your implementation
}

func (c *CustomLogger) Info(msg string, args ...interface{}) {
    // Your implementation
}

func (c *CustomLogger) Warn(msg string, args ...interface{}) {
    // Your implementation
}

func (c *CustomLogger) Error(msg string, args ...interface{}) {
    // Your implementation
}

func (c *CustomLogger) SetLevel(level logger.LogLevel) {
    // Your implementation
}

func (c *CustomLogger) GetLevel() logger.LogLevel {
    // Your implementation
}
```

## Examples

### Example 1: Simple Usage

```go
package main

import "github.com/snipwise/snip-sdk/snip/toolbox/logger"

func main() {
    log := logger.NewConsoleLogger(logger.LevelInfo)

    log.Debug("This won't show (level is INFO)")
    log.Info("Application started")
    log.Warn("Using default configuration")
    log.Error("Failed to load config: %s", "file not found")
}
```

### Example 2: With Environment Variable

```bash
# Set environment variable
export SNIP_LOG_LEVEL=debug

# Run your application
go run main.go
```

```go
package main

import "github.com/snipwise/snip-sdk/snip/toolbox/logger"

func main() {
    // Logger will use SNIP_LOG_LEVEL environment variable
    log := logger.GetLoggerFromEnv()

    log.Debug("Debug information")
    log.Info("Info message")
}
```

### Example 3: Agent Integration

```go
// Set log level via environment
// export SNIP_LOG_LEVEL=info

agent, err := chat.NewChatAgent(
    ctx,
    agentConfig,
    modelConfig,
    chat.WithLogger(logger.GetLoggerFromEnvWithPrefix("ChatAgent")),
)

// Or use verbose mode option
agent, err := chat.NewChatAgent(
    ctx,
    agentConfig,
    modelConfig,
    chat.WithVerbose(true), // Enables INFO level logging
)
```

## Best Practices

1. **Use environment variables for global control**: Let users control logging without code changes
2. **Use prefixes for components**: Makes it easier to identify log sources
3. **Default to NoOp**: Don't produce output unless explicitly requested
4. **Use appropriate levels**:
   - DEBUG: Detailed internal state
   - INFO: Important operations and confirmations
   - WARN: Potential issues that don't prevent operation
   - ERROR: Actual errors that need attention
5. **Format messages clearly**: Include context and relevant data
