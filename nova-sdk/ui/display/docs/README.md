# Display Package Documentation

Welcome to the Nova SDK Display package documentation. This package provides comprehensive utilities for creating rich, colorful terminal output including formatted text, markdown rendering, and streaming markdown support.

## Overview

The display package offers three main components:

1. **Display Functions** - Rich terminal output with colors, styles, and semantic formatting
2. **Markdown Renderer** - Full markdown rendering for static content
3. **Streaming Markdown Parser** - Real-time markdown rendering for streaming content

All components work together to provide a consistent, beautiful terminal experience.

## Documentation Files

### Core Documentation

- **[display.md](display.md)** - Complete reference for display functions
  - Color and style constants
  - Semantic message functions (Success, Error, Warning, Info, Debug)
  - Headers, titles, and separators
  - Lists, bullets, and numbered items
  - Boxes, banners, and highlights
  - Tables and key-value pairs
  - Progress indicators and step counters
  - Indentation and structured output
  - Object-like output formatting

### Markdown Documentation

- **[markdown.md](markdown.md)** - Static markdown rendering
  - Complete markdown syntax support
  - Headers, code blocks, lists
  - Inline formatting (bold, italic, code)
  - Links, images, blockquotes
  - Task lists and horizontal rules
  - Color scheme and rendering details

- **[markdown.chunk.md](markdown.chunk.md)** - Streaming markdown parser
  - Real-time markdown rendering
  - Stateful chunk processing
  - Smart buffering and detection
  - AI response streaming
  - HTTP streaming support
  - Performance characteristics

## Quick Start

### Basic Display Functions

```go
import "your-module/display"

// Semantic messages
display.Success("Operation completed")
display.Error("Something went wrong")
display.Warning("Be careful")
display.Info("Just so you know")

// Styled output
display.Header("Main Section")
display.Subheader("Subsection")
display.Bullet("First point")
display.Bullet("Second point")
```

### Static Markdown

```go
markdown := `
# Documentation

## Features

- Easy to use
- **Powerful** rendering
- \`Code\` support

\`\`\`go
func main() {
    fmt.Println("Hello")
}
\`\`\`
`

display.Markdown(markdown)
```

### Streaming Markdown

```go
parser := display.NewMarkdownChunkParser()
defer parser.Flush()

for chunk := range streamingSource {
    display.MarkdownChunk(parser, chunk)
}
```

## Feature Comparison

### Display Functions vs Markdown

| Feature | Display Functions | Markdown |
|---------|------------------|----------|
| **Use Case** | Programmatic output | Document rendering |
| **Flexibility** | High - build dynamically | Medium - pre-formatted |
| **Performance** | Fast - direct output | Medium - parsing overhead |
| **Formatting** | Explicit function calls | Markdown syntax |
| **Best For** | Status messages, logs | Documentation, AI responses |

### Static vs Streaming Markdown

| Feature | Static (`Markdown`) | Streaming (`MarkdownChunk`) |
|---------|--------------------|-----------------------------|
| **Input** | Complete string | Incremental chunks |
| **Display** | After complete | Real-time |
| **Memory** | Full document | Current line only |
| **Use Case** | Files, strings | AI, HTTP streams |
| **Latency** | Higher | Lower (immediate) |
| **Complexity** | Simple | Stateful |

## Common Use Cases

### 1. CLI Application Output

```go
display.Title("My Application")
display.Info("Starting initialization...")

display.Step(1, 3, "Loading configuration")
// ... load config ...
display.Done("Configuration loaded")

display.Step(2, 3, "Connecting to database")
// ... connect ...
display.Success("Connected to database")

display.Step(3, 3, "Starting server")
// ... start server ...
display.Success("Server running on port 8080")
```

### 2. Error Reporting

```go
if err != nil {
    display.Error("Failed to process request")
    display.Indent(1, fmt.Sprintf("Reason: %v", err))
    display.Indent(1, "Please check your configuration")
    return
}
```

### 3. Configuration Display

```go
display.Header("Current Configuration")
display.Table("Environment", config.Env)
display.Table("Port", fmt.Sprintf("%d", config.Port))
display.Table("Database", config.DatabaseURL)
display.Separator()
```

### 4. Progress Tracking

```go
display.Progress("Processing items")
for i, item := range items {
    display.Step(i+1, len(items), fmt.Sprintf("Processing %s", item.Name))
    // ... process item ...
}
display.Done("All items processed")
```

### 5. AI Response Streaming

Example of how to integrate with an AI streaming API:

```go
// Example integration (pseudocode - adapt to your AI client)
func streamAIResponse(client *ai.Client, prompt string) {
    parser := display.NewMarkdownChunkParser()
    defer parser.Flush()

    stream, err := client.StreamCompletion(prompt)
    if err != nil {
        display.Error(fmt.Sprintf("Stream error: %v", err))
        return
    }

    for chunk := range stream {
        display.MarkdownChunk(parser, chunk.Text)
    }
}
```

### 6. Documentation Rendering

Example of using Markdown for help text:

```go
func showHelp() {
    helpText := `
# Application Help

## Commands

- \`start\` - Start the server
- \`stop\` - Stop the server
- \`status\` - Check server status

## Options

Use \`--help\` with any command for details.
    `
    display.Markdown(helpText)
}
```

### 7. Structured Data Display

```go
display.ObjectStart("User")
display.Field("id", user.ID)
display.Field("name", user.Name)
display.Field("email", user.Email)
display.Field("role", user.Role)
display.ObjectEnd()
```

### 8. Multi-step Wizard

```go
display.Banner("Setup Wizard")

display.Header("Step 1: Database Configuration")
// ... get database config ...
display.Success("Database configured")
display.NewLine()

display.Header("Step 2: Authentication")
// ... configure auth ...
display.Success("Authentication configured")
display.NewLine()

display.Header("Step 3: Complete")
display.Done("Setup complete!")
```

## Color Reference

### Standard Colors

- **Red**: Errors, critical alerts
- **Green**: Success, completion
- **Yellow**: Warnings, progress
- **Cyan**: Info, headers, links
- **Blue**: Subheaders, secondary info
- **Magenta**: Special highlights
- **Gray**: Debug, secondary text

### Best Practices

1. **Consistency**: Use semantic functions (Success, Error, etc.) for consistent messaging
2. **Hierarchy**: Use headers and indentation to show structure
3. **Emphasis**: Use bold/bright colors sparingly for important information
4. **Readability**: Add separators and whitespace between sections
5. **Progressive**: Show progress for long operations

## Terminal Compatibility

### Fully Supported

- macOS Terminal
- iTerm2
- Linux terminals (GNOME Terminal, Konsole, etc.)
- Windows Terminal
- PowerShell 7+
- VSCode integrated terminal
- Most modern CI/CD environments

### Partial Support

- Windows Command Prompt (limited colors)
- Basic SSH terminals
- Older terminal emulators

### Graceful Degradation

The package uses ANSI escape codes that gracefully degrade:
- If colors aren't supported, text still displays correctly
- Unicode symbols fall back to basic characters in some terminals
- Core functionality works everywhere

## Advanced Features

### Custom Styling

```go
// Combine multiple styles
display.Styled("ALERT", ColorBold, ColorRed, BgYellow)

// Custom highlights
display.Highlight(" Critical ", ColorWhite, BgRed)

// Complex formatting
display.Styledf("Status: %s", []string{ColorBold, ColorGreen}, status)
```

### Dynamic Content

```go
// Build content programmatically
for i, item := range items {
    if item.IsComplete {
        display.ColoredList(i+1, item.Name, ColorGreen)
    } else {
        display.ColoredList(i+1, item.Name, ColorGray)
    }
}
```

### Nested Structures

```go
display.Header("Configuration")
display.Bullet("Database")
display.Indent(1, "Host: localhost")
display.Indent(1, "Port: 5432")
display.Bullet("Cache")
display.Indent(1, "Type: Redis")
display.Indent(1, "TTL: 3600s")
```

## Performance Tips

### For Static Content

```go
// Good: Single Markdown call
display.Markdown(entireDocument)

// Avoid: Multiple small calls
for _, line := range lines {
    display.Markdown(line) // Creates parsing overhead
}
```

### For Streaming Content

```go
// Good: Reuse parser
parser := display.NewMarkdownChunkParser()
for doc := range documents {
    for chunk := range doc {
        display.MarkdownChunk(parser, chunk)
    }
    parser.Flush()
    parser.Reset()
}

// Avoid: New parser each time
for chunk := range stream {
    p := display.NewMarkdownChunkParser() // Wasteful
    display.MarkdownChunk(p, chunk)
}
```

### For Repeated Output

```go
// Good: Pre-format color codes
successPrefix := fmt.Sprintf("%s%s ", ColorGreen, SymbolSuccess)
for _, item := range items {
    fmt.Printf("%s%s%s\n", successPrefix, item, ColorReset)
}

// Avoid: Repeated function calls
for _, item := range items {
    display.Success(item) // Function call overhead
}
```

## Package Structure

```
display/
├── display.go           # Core display functions
├── markdown.go          # Static markdown renderer
├── markdown.chunk.go    # Streaming markdown parser
└── docs/
    ├── README.md        # This file
    ├── display.md       # Display functions reference
    ├── markdown.md      # Static markdown docs
    └── markdown.chunk.md # Streaming markdown docs
```

## API Categories

### Message Functions
- Success, Error, Warning, Info, Debug
- Progress, Done
- Variants with formatting (Successf, Errorf, etc.)

### Structural Functions
- Header, Subheader, Title
- Separator, Box, Banner
- Bullet, List, Arrow

### Data Display
- Table, KeyValue
- ObjectStart, Field, ObjectEnd
- Indent, ColoredIndent

### Styling Functions
- Bold, Italic, Underline
- Color, Colorln, Colorf
- Styled, Highlight

### Markdown Functions
- Markdown (static)
- NewMarkdownChunkParser, MarkdownChunk (streaming)

### Utility Functions
- Print, Println, Printf
- NewLine, Clear
- Step, Indent

## Examples Repository

For more comprehensive examples, see:
- Basic examples in each documentation file
- Real-world use cases in this README
- Integration examples in the main Nova SDK documentation

## Contributing

When adding new display functions:
1. Follow existing naming conventions
2. Provide both base and formatted variants (e.g., `Func` and `Funcf`)
3. Use package color constants
4. Always reset colors after output
5. Document in the appropriate file

## License

Part of the Nova SDK project.
