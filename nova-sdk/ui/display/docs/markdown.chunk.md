# Markdown Chunk Parser Documentation

## Overview

The `markdown.chunk.go` file provides a streaming markdown parser that can process and render markdown content incrementally as it arrives. This is particularly useful for real-time applications like streaming AI responses, live data feeds, or progressive document rendering.

## File: markdown.chunk.go

### Main Components

#### MarkdownChunkParser

A stateful parser that maintains context while processing markdown chunks.

**Structure:**
```go
type MarkdownChunkParser struct {
    buffer          strings.Builder  // Accumulates incomplete chunks
    inCodeBlock     bool              // Currently inside code block
    codeBlockLang   string            // Language of current code block
    lineBuffer      string            // Current line being processed
    lastWasNewline  bool              // Track newline state
    inList          bool              // Currently inside a list
    alreadyPrinted  bool              // Line already displayed during streaming
}
```

**Features:**
- Stateful parsing for streaming content
- Line-by-line processing
- Real-time display for plain text
- Buffering for incomplete lines
- Proper handling of markdown structures

### Constructor

#### NewMarkdownChunkParser

```go
func NewMarkdownChunkParser() *MarkdownChunkParser
```

Creates a new markdown chunk parser instance.

**Returns:**
- `*MarkdownChunkParser` - New parser with initialized state

**Example:**
```go
parser := display.NewMarkdownChunkParser()
```

### Main Functions

#### MarkdownChunk

```go
func MarkdownChunk(parser *MarkdownChunkParser, chunk string)
```

Processes and displays a chunk of markdown text with streaming support. This is the main entry point for streaming markdown rendering.

**Parameters:**
- `parser` - Parser instance maintaining state
- `chunk` - Incoming chunk of markdown text

**Behavior:**
1. Appends chunk to internal buffer
2. Splits content by newlines
3. Processes all complete lines immediately
4. Keeps incomplete last line in buffer
5. Displays plain text chunks immediately for responsive output
6. Waits for complete lines before rendering markdown structures

**Smart Streaming:**
The function intelligently determines when to display content immediately:
- Plain text is streamed character-by-character for responsiveness
- Markdown structures (headers, lists, etc.) wait for complete lines
- Ambiguous content (like single digits that might become "1.") waits for clarification

**Example:**
```go
parser := display.NewMarkdownChunkParser()

// Simulate streaming chunks
chunks := []string{
    "# Hello",
    " World\n\n",
    "This is **bold",
    "** text.\n",
    "- Item 1\n",
    "- Item 2\n",
}

for _, chunk := range chunks {
    display.MarkdownChunk(parser, chunk)
}

parser.Flush() // Don't forget to flush at the end
```

#### Flush

```go
func (p *MarkdownChunkParser) Flush()
```

Processes any remaining buffered content. **Must be called** at the end of streaming to ensure all content is displayed.

**Example:**
```go
parser := display.NewMarkdownChunkParser()
// ... process chunks ...
parser.Flush() // Ensure all buffered content is displayed
```

#### Reset

```go
func (p *MarkdownChunkParser) Reset()
```

Resets the parser state to initial values. Useful for reusing the parser for a new document.

**Example:**
```go
parser.Reset() // Clear all state and buffers
// Parser is now ready for new content
```

### Internal Functions

#### processLine

```go
func (p *MarkdownChunkParser) processLine()
```

Internal function that processes a complete line of markdown. Handles all markdown syntax recognition and rendering.

**Processes:**
- Code block delimiters (\`\`\`)
- Code block content
- Headers (#, ##, ###, etc.)
- Horizontal rules (---, ***, ___)
- Blockquotes (>)
- Unordered lists (-, *, +)
- Ordered lists (1., 2., 3., etc.)
- Task lists (- [ ], - [x])
- Regular paragraphs with inline formatting

**Not directly called by users** - invoked internally by `MarkdownChunk`.

#### eraseAndPrint

```go
func (p *MarkdownChunkParser) eraseAndPrint(formatted string)
```

Internal function that prints formatted content.

**Parameters:**
- `formatted` - Formatted text with ANSI codes

**Purpose:**
- Handles output of processed markdown lines
- Manages line clearing for updates

**Not directly called by users** - invoked internally by `processLine`.

#### tryProcessInline

```go
func (p *MarkdownChunkParser) tryProcessInline()
```

Internal placeholder function kept for compatibility. Currently performs no operations as character streaming is handled directly in `MarkdownChunk`.

## Streaming Behavior

### Smart Detection

The parser intelligently detects whether incoming text is:

1. **Plain Text**: Displayed immediately for responsive output
2. **Markdown Structure Start**: Buffered until complete
3. **Ambiguous Content**: Buffered until clarified

**Detected Markdown Structures:**
- Headers: `#`, `##`, `###`, etc.
- Code blocks: \`\`\`
- Blockquotes: `>`
- List markers: `-`, `*`, `+`
- Ordered lists: `1.`, `2.`, etc.
- Horizontal rules: `---`, `***`, `___`

**Ambiguous Patterns:**
- Single digits (might become `1.` for ordered list)
- Empty lines (might separate blocks)

### Line Buffering

The parser maintains a line buffer to handle incomplete lines:

```
Chunk 1: "# Hello Wo"    → Buffered (incomplete header)
Chunk 2: "rld\n"         → Processed as "# Hello World" (complete)
Chunk 3: "Plain te"      → Displayed immediately (plain text)
Chunk 4: "xt\n"          → Newline added, line complete
```

### State Persistence

The parser maintains state across chunks:

- **inCodeBlock**: Ensures code content is rendered correctly
- **codeBlockLang**: Remembers the language for display
- **inList**: Maintains list context for continuation lines
- **alreadyPrinted**: Prevents duplicate output

## Usage Examples

### Basic Streaming

```go
parser := display.NewMarkdownChunkParser()

// Simulate receiving chunks from a stream
stream := []string{
    "# Streaming",
    " Markdown\n\n",
    "This text arrives ",
    "piece by piece.\n\n",
}

for _, chunk := range stream {
    display.MarkdownChunk(parser, chunk)
    time.Sleep(100 * time.Millisecond) // Simulate delay
}

parser.Flush()
```

### AI Response Streaming Example

Here's an example of how you could use the chunk parser with AI streaming:

```go
// Example user function (not part of the display package)
func streamAIResponse(responseChan <-chan string) {
    parser := display.NewMarkdownChunkParser()

    for chunk := range responseChan {
        display.MarkdownChunk(parser, chunk)
    }

    parser.Flush()
}
```

### HTTP Streaming Example

Here's an example of how you could use the chunk parser with HTTP streaming:

```go
// Example user function (not part of the display package)
func handleStreamingResponse(resp *http.Response) {
    parser := display.NewMarkdownChunkParser()
    defer parser.Flush()

    reader := bufio.NewReader(resp.Body)
    buffer := make([]byte, 1024)

    for {
        n, err := reader.Read(buffer)
        if n > 0 {
            chunk := string(buffer[:n])
            display.MarkdownChunk(parser, chunk)
        }
        if err != nil {
            break
        }
    }
}
```

### Parser Reuse

```go
parser := display.NewMarkdownChunkParser()

// Process first document
for _, chunk := range document1 {
    display.MarkdownChunk(parser, chunk)
}
parser.Flush()

// Reset and process second document
parser.Reset()
for _, chunk := range document2 {
    display.MarkdownChunk(parser, chunk)
}
parser.Flush()
```

### Code Block Streaming

```go
parser := display.NewMarkdownChunkParser()

chunks := []string{
    "```go\n",
    "func main() {\n",
    "    fmt.Println(",
    "\"Hello",
    "\")\n",
    "}\n",
    "```\n",
}

for _, chunk := range chunks {
    display.MarkdownChunk(parser, chunk)
}

parser.Flush()
```

Output:
```
┌─ Code: go
│ func main() {
│     fmt.Println("Hello")
│ }
└─
```

## State Management

### Parser States

The parser tracks several states:

| State | Type | Purpose |
|-------|------|---------|
| `buffer` | `strings.Builder` | Accumulates incomplete chunks |
| `inCodeBlock` | `bool` | Inside code block fence |
| `codeBlockLang` | `string` | Current code block language |
| `lineBuffer` | `string` | Current line being processed |
| `lastWasNewline` | `bool` | Track line boundaries |
| `inList` | `bool` | Inside list structure |
| `alreadyPrinted` | `bool` | Prevent duplicate printing |

### State Transitions

**Entering Code Block:**
```
Normal → Sees ``` → Sets inCodeBlock=true → Stores language
```

**Exiting Code Block:**
```
In Code Block → Sees ``` → Sets inCodeBlock=false → Clears language
```

**List Detection:**
```
Normal → Sees list marker → Sets inList=true
```

**List Exit:**
```
In List → Sees empty line → Sets inList=false
```

## Best Practices

### 1. Always Flush

```go
parser := display.NewMarkdownChunkParser()
defer parser.Flush() // Ensure cleanup

for chunk := range stream {
    display.MarkdownChunk(parser, chunk)
}
```

### 2. Reuse Parsers

```go
// Good: Reuse parser with Reset
parser := display.NewMarkdownChunkParser()
for _, doc := range documents {
    for _, chunk := range doc {
        display.MarkdownChunk(parser, chunk)
    }
    parser.Flush()
    parser.Reset()
}

// Avoid: Creating new parser each time
for _, doc := range documents {
    parser := display.NewMarkdownChunkParser() // Wasteful
    // ...
}
```

### 3. Handle Errors Gracefully

```go
parser := display.NewMarkdownChunkParser()
defer parser.Flush()

for chunk := range stream {
    if chunk == "" {
        continue // Skip empty chunks
    }
    display.MarkdownChunk(parser, chunk)
}
```

### 4. Chunk Size Considerations

```go
// Good: Reasonable chunk sizes (1-4KB)
buffer := make([]byte, 2048)

// Avoid: Too small (overhead) or too large (latency)
// buffer := make([]byte, 10) // Too small
// buffer := make([]byte, 1048576) // Too large for streaming
```

## Integration with Display Package

The chunk parser uses the same rendering functions as the static markdown renderer:

- Uses `formatInlineMarkdown()` for inline formatting
- Uses `countLeadingSpaces()` for indentation
- Applies same color scheme and styling
- Integrates with display package constants

This ensures consistent output between static and streaming rendering.

## Performance Characteristics

### Memory Usage

- **Buffer**: Grows only to size of longest incomplete line
- **State**: Fixed size (~100 bytes)
- **Total**: Minimal memory footprint

### Processing Speed

- **Per Chunk**: O(n) where n is chunk size
- **Line Processing**: O(m) where m is line length
- **Regex**: Compiled once, reused for all lines

### Latency

- Plain text: **Immediate** display
- Markdown structures: Display after newline (minimal buffering)
- Code blocks: Line-by-line display within block

## Limitations

### Current Limitations

1. **No Look-Ahead**: Cannot handle markdown syntax split across chunks mid-token
2. **Line-Based**: Assumes line breaks in markdown source
3. **No Backtracking**: Once displayed, cannot modify previous output
4. **Limited Rollback**: Cannot undo printed characters

### Workarounds

**For Split Tokens:**
```go
// If source might have "**bo" + "ld**", ensure chunks preserve tokens
// Or use larger buffer sizes to reduce split probability
```

**For Complex Structures:**
```go
// For tables or complex layouts, buffer the entire structure
// before passing to chunk parser
```

## Differences from Static Markdown

| Feature | Static (`Markdown`) | Streaming (`MarkdownChunk`) |
|---------|--------------------|-----------------------------|
| Processing | All at once | Incremental |
| Display | After complete | Real-time |
| Memory | Full document | Current line only |
| State | Local variables | Persistent object |
| Use Case | Complete documents | Streaming data |
| Latency | Higher | Lower |

## Thread Safety

**Not Thread-Safe**: The parser maintains internal state and is not safe for concurrent use. Use separate parser instances for different goroutines.

```go
// Good: One parser per goroutine
go func() {
    parser := display.NewMarkdownChunkParser()
    // ... use parser ...
}()

// Bad: Sharing parser across goroutines
parser := display.NewMarkdownChunkParser()
go func() { display.MarkdownChunk(parser, chunk1) }() // Race condition!
go func() { display.MarkdownChunk(parser, chunk2) }() // Race condition!
```
