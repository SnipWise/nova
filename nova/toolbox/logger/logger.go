package logger

/*
// 1. Via environment variable (recommended)
export SNIP_LOG_LEVEL=info
agent, _ := chat.NewChatAgent(ctx, agentConfig, modelConfig)

// 2. Via WithVerbose option
agent, _ := chat.NewChatAgent(ctx, agentConfig, modelConfig,
    chat.WithVerbose(true))

// 3. Via WithLogLevel option
agent, _ := chat.NewChatAgent(ctx, agentConfig, modelConfig,
    chat.WithLogLevel(logger.LevelDebug))

// 4. Via WithLogger option (custom logger)
agent, _ := chat.NewChatAgent(ctx, agentConfig, modelConfig,
    chat.WithLogger(customLogger))
*/


import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// LevelDebug is for detailed debugging information
	LevelDebug LogLevel = iota
	// LevelInfo is for informational messages
	LevelInfo
	// LevelWarn is for warning messages
	LevelWarn
	// LevelError is for error messages
	LevelError
	// LevelNone disables all logging
	LevelNone
)

// String returns the string representation of a LogLevel
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelNone:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel converts a string to a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "NONE":
		return LevelNone
	default:
		return LevelInfo
	}
}

// Logger is the interface for logging in the snip-sdk
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, args ...interface{})
	// Info logs an informational message
	Info(msg string, args ...interface{})
	// Warn logs a warning message
	Warn(msg string, args ...interface{})
	// Error logs an error message
	Error(msg string, args ...interface{})
	// SetLevel sets the minimum log level
	SetLevel(level LogLevel)
	// GetLevel returns the current log level
	GetLevel() LogLevel
}

// NoOpLogger is a logger that does nothing (default)
type NoOpLogger struct{}

// Debug does nothing
func (n *NoOpLogger) Debug(msg string, args ...interface{}) {}

// Info does nothing
func (n *NoOpLogger) Info(msg string, args ...interface{}) {}

// Warn does nothing
func (n *NoOpLogger) Warn(msg string, args ...interface{}) {}

// Error does nothing
func (n *NoOpLogger) Error(msg string, args ...interface{}) {}

// SetLevel does nothing
func (n *NoOpLogger) SetLevel(level LogLevel) {}

// GetLevel returns LevelNone
func (n *NoOpLogger) GetLevel() LogLevel {
	return LevelNone
}

// ConsoleLogger logs to stdout/stderr
type ConsoleLogger struct {
	level  LogLevel
	prefix string
}

// NewConsoleLogger creates a new console logger with the specified level
func NewConsoleLogger(level LogLevel) *ConsoleLogger {
	return &ConsoleLogger{
		level: level,
	}
}

// NewConsoleLoggerWithPrefix creates a new console logger with a prefix
func NewConsoleLoggerWithPrefix(level LogLevel, prefix string) *ConsoleLogger {
	return &ConsoleLogger{
		level:  level,
		prefix: prefix,
	}
}

// SetLevel sets the minimum log level
func (c *ConsoleLogger) SetLevel(level LogLevel) {
	c.level = level
}

// GetLevel returns the current log level
func (c *ConsoleLogger) GetLevel() LogLevel {
	return c.level
}

// Debug logs a debug message
func (c *ConsoleLogger) Debug(msg string, args ...interface{}) {
	if c.level <= LevelDebug {
		c.log("DEBUG", msg, args...)
	}
}

// Info logs an informational message
func (c *ConsoleLogger) Info(msg string, args ...interface{}) {
	if c.level <= LevelInfo {
		c.log("INFO", msg, args...)
	}
}

// Warn logs a warning message
func (c *ConsoleLogger) Warn(msg string, args ...interface{}) {
	if c.level <= LevelWarn {
		c.log("WARN", msg, args...)
	}
}

// Error logs an error message
func (c *ConsoleLogger) Error(msg string, args ...interface{}) {
	if c.level <= LevelError {
		c.log("ERROR", msg, args...)
	}
}

// log is the internal logging function
func (c *ConsoleLogger) log(level string, msg string, args ...interface{}) {
	prefix := ""
	if c.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", c.prefix)
	}

	// Format the message with args if provided
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	log.Printf("%s[%s] %s", prefix, level, formattedMsg)
}

// GetLoggerFromEnv creates a logger based on environment variables
// SNIP_LOG_LEVEL: debug, info, warn, error, none (default: none)
// Returns a NoOpLogger if SNIP_LOG_LEVEL is not set or set to "none"
func GetLoggerFromEnv() Logger {
	level := os.Getenv("SNIP_LOG_LEVEL")
	if level == "" {
		return &NoOpLogger{}
	}

	logLevel := ParseLogLevel(level)
	if logLevel == LevelNone {
		return &NoOpLogger{}
	}

	return NewConsoleLogger(logLevel)
}

// GetLoggerFromEnvWithPrefix creates a logger based on environment variables with a prefix
func GetLoggerFromEnvWithPrefix(prefix string) Logger {
	level := os.Getenv("SNIP_LOG_LEVEL")
	if level == "" {
		return &NoOpLogger{}
	}

	logLevel := ParseLogLevel(level)
	if logLevel == LevelNone {
		return &NoOpLogger{}
	}

	return NewConsoleLoggerWithPrefix(logLevel, prefix)
}
