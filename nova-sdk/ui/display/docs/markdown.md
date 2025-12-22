# Markdown Rendering Documentation

## Overview

The `markdown.go` file provides a complete markdown rendering engine that converts markdown-formatted text into colored, styled terminal output. It supports most common markdown features including headers, lists, code blocks, inline formatting, and more.

## File: markdown.go

### Main Function

#### Markdown

```go
func Markdown(content string)
```

Renders and prints formatted markdown content with ANSI colors and styling to the terminal.

**Parameters:**
- `content` - Markdown-formatted string to render

**Features:**
- Full markdown syntax support
- Colored output with semantic highlighting
- Code block rendering with language detection
- Inline formatting (bold, italic, code, links, etc.)
- Lists (ordered, unordered, task lists)
- Headers with decorative underlines
- Blockquotes
- Horizontal rules

**Example:**
```go
markdown := `
# Main Title

This is a **bold** statement with *italic* text and \`inline code\`.

## Subtitle

- First item
- Second item
  - Nested item

\`\`\`go
func main() {
    fmt.Println("Hello, World!")
}
\`\`\`
`

display.Markdown(markdown)
```

## Supported Markdown Features

### Headers

Supports 6 levels of headers using `#` syntax:

**Syntax:**
```markdown
# Level 1 Header
## Level 2 Header
### Level 3 Header
#### Level 4 Header
```

**Rendering:**
- **Level 1**: Bold bright cyan with double-line underline (═)
- **Level 2**: Bold bright blue with single-line underline (─)
- **Level 3**: Bold cyan (no underline)
- **Level 4+**: Bright cyan (no underline)

### Code Blocks

Fenced code blocks with optional language specification:

**Syntax:**
```markdown
\`\`\`python
def hello():
    print("Hello")
\`\`\`
```

**Rendering:**
- Gray bordered box with language label
- Original indentation preserved
- Syntax: `┌─ Code: [language]` at top, `└─` at bottom
- Code lines prefixed with `│`

### Inline Code

Inline code using backticks:

**Syntax:**
```markdown
Use the \`print()\` function
```

**Rendering:**
- Bright yellow text
- Black background
- Automatically reset after code

### Text Formatting

#### Bold

**Syntax:**
```markdown
**bold text**
__also bold__
```

**Rendering:**
- Bold ANSI style applied

#### Italic

**Syntax:**
```markdown
*italic text*
_also italic_
```

**Rendering:**
- Italic ANSI style applied

#### Strikethrough

**Syntax:**
```markdown
~~strikethrough text~~
```

**Rendering:**
- Dim ANSI style applied

#### Combined Formatting

**Syntax:**
```markdown
**\`bold code\`**
```

**Rendering:**
- Bold + bright yellow + black background

### Lists

#### Unordered Lists

**Syntax:**
```markdown
- First item
* Second item
+ Third item
  - Nested item
```

**Rendering:**
- Bright yellow bullet symbol (•)
- Supports nested lists with proper indentation (2 spaces per level)
- Inline formatting supported within list items

#### Ordered Lists

**Syntax:**
```markdown
1. First item
2. Second item
3. Third item
   1. Nested item
```

**Rendering:**
- Bright yellow numbers with period
- Supports nested lists with proper indentation
- Inline formatting supported within list items

#### Task Lists

**Syntax:**
```markdown
- [ ] Unchecked task
- [x] Checked task
- [X] Also checked
```

**Rendering:**
- Unchecked: Gray checkbox (☐)
- Checked: Green checkmark (✓)
- Inline formatting supported within task text

### Blockquotes

**Syntax:**
```markdown
> This is a quoted text
> Multiple lines supported
```

**Rendering:**
- Gray vertical bar (│) prefix
- Inline formatting supported within quotes

### Horizontal Rules

**Syntax:**
```markdown
---
***
___
```

**Rendering:**
- 80-character line of dashes (─)

### Links

**Syntax:**
```markdown
[Link text](https://example.com)
```

**Rendering:**
- Bright cyan underlined text for link text
- Gray URL in parentheses after link text
- Format: `Link text (https://example.com)`

### Images

**Syntax:**
```markdown
![Alt text](image.png)
```

**Rendering:**
- Bright magenta with picture emoji (🖼)
- Gray URL in parentheses
- Format: `🖼  Alt text (image.png)`

### Escaped Characters

The parser supports escaping special markdown characters:

**Syntax:**
```markdown
\* Not a list
\_ Not italic
\` Not code
```

**Rendering:**
- Escaped characters displayed literally

## Helper Functions

### formatInlineMarkdown

```go
func formatInlineMarkdown(text string) string
```

Internal function that processes inline markdown formatting within a line of text.

**Processing Order:**
1. Escape character handling
2. Bold + code combinations (`**\`text\`**`)
3. Bold text (`**text**` or `__text__`)
4. Inline code (\`text\`)
5. Italic text (`*text*` or `_text_`)
6. Strikethrough (`~~text~~`)
7. Links (`[text](url)`)
8. Images (`![alt](url)`)
9. Restore escaped characters

**Parameters:**
- `text` - Text with inline markdown syntax

**Returns:**
- Formatted text with ANSI color codes

### countLeadingSpaces

```go
func countLeadingSpaces(line string) int
```

Internal helper function that counts leading spaces in a line.

**Parameters:**
- `line` - Line of text

**Returns:**
- Number of leading spaces (used for list indentation)

## Rendering Details

### Color Scheme

| Element | Foreground Color | Background | Style |
|---------|-----------------|------------|-------|
| H1 Header | Bright Cyan | None | Bold |
| H2 Header | Bright Blue | None | Bold |
| H3 Header | Cyan | None | Bold |
| Code Block Border | Gray | None | Dim |
| Code Text | Default | None | None |
| Inline Code | Bright Yellow | Black | None |
| Bold | Default | None | Bold |
| Italic | Default | None | Italic |
| Strikethrough | Default | None | Dim |
| List Bullet/Number | Bright Yellow | None | None |
| Task Checked | Green | None | None |
| Task Unchecked | Gray | None | None |
| Blockquote | Gray | None | None |
| Link Text | Bright Cyan | None | Underline |
| Link URL | Gray | None | None |
| Image | Bright Magenta | None | None |

### State Management

The `Markdown` function maintains state while parsing:
- `inCodeBlock` - Tracks whether currently inside a code block
- `inList` - Tracks whether currently inside a list
- `listLevel` - Current list nesting level

This state helps with proper formatting of continuation lines and nested structures.

## Usage Examples

### Complete Document

```go
content := `
# Project Documentation

## Overview

This is a **powerful** markdown renderer with *full* support for:

- Code blocks
- Inline formatting
- Lists and more

## Installation

\`\`\`bash
go get github.com/example/display
\`\`\`

## Usage

Use the \`Markdown()\` function:

\`\`\`go
display.Markdown(content)
\`\`\`

> **Note**: Remember to import the package first!

### Features

1. Headers (6 levels)
2. **Bold** and *italic* text
3. \`Inline code\`
4. [Links](https://example.com)

---

**Happy coding!** 🎉
`

display.Markdown(content)
```

### Code Documentation

```go
codeDoc := `
## Function: ProcessData

Processes incoming data and returns results.

### Parameters

- \`data\` - Input data to process
- \`options\` - Processing options

### Example

\`\`\`go
result, err := ProcessData(data, options)
if err != nil {
    log.Fatal(err)
}
\`\`\`

### Returns

Returns processed data or error.
`

display.Markdown(codeDoc)
```

### Task List

```go
tasks := `
## Today's Tasks

- [x] Write documentation
- [x] Review pull requests
- [ ] Deploy to production
- [ ] Update changelog

### Next Steps

1. Test all features
2. Get peer review
3. Merge changes
`

display.Markdown(tasks)
```

## Limitations and Notes

### Current Limitations

1. **Tables**: Not currently supported
2. **Nested Blockquotes**: Limited support
3. **HTML**: Raw HTML is not parsed or rendered
4. **Reference Links**: Not supported (only inline links)
5. **Footnotes**: Not supported

### Best Practices

1. **Line Breaks**: Use blank lines between different elements for best rendering
2. **Code Blocks**: Always specify language for better visual clarity
3. **Indentation**: Use consistent indentation (2 or 4 spaces) for nested lists
4. **Escaping**: Use backslash to escape markdown special characters when needed

### Terminal Compatibility

The markdown renderer uses Unicode characters and ANSI color codes:
- Requires UTF-8 terminal support for symbols
- Works best with modern terminal emulators
- May have reduced visual quality in basic terminals
- Colors can be disabled by terminal settings

## Performance Considerations

- The function processes the entire content string at once
- For very large documents, consider splitting into sections
- Regex operations are used for pattern matching (reasonable performance)
- No caching - each call reprocesses the entire content

## Integration with Display Package

The `Markdown` function uses color constants and symbols from the main display package:
- `ColorBold`, `ColorItalic`, `ColorDim`
- `ColorBrightCyan`, `ColorBrightBlue`, etc.
- `ColorReset` for proper color boundaries
- `SymbolBullet`, `SymbolCheck` for list rendering

This ensures consistent visual style across all display package functions.
