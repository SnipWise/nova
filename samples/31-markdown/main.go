package main

import (
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Example markdown content
	markdownContent := "# Welcome to Markdown Display\n\n" +
		"This is a **demo** of the markdown rendering capabilities.\n\n" +
		"## Features\n\n" +
		"The markdown parser supports various elements:\n\n" +
		"### Inline Formatting\n\n" +
		"You can use **bold text**, *italic text*, ~~strikethrough~~, and `inline code`.\n\n" +
		"### Lists\n\n" +
		"Unordered lists:\n" +
		"- First item with **bold**\n" +
		"- Second item with *italic*\n" +
		"  - Nested item 1\n" +
		"  - Nested item 2\n" +
		"- Third item with `code`\n\n" +
		"Ordered lists:\n" +
		"1. First step\n" +
		"2. Second step\n" +
		"3. Third step with [a link](https://example.com)\n\n" +
		"### Task Lists\n\n" +
		"- [x] Completed task\n" +
		"- [ ] Pending task\n" +
		"- [x] Another completed task\n\n" +
		"### Code Blocks\n\n" +
		"```go\n" +
		"package main\n\n" +
		"func main() {\n" +
		"    fmt.Println(\"Hello, World!\")\n" +
		"}\n" +
		"```\n\n" +
		"```javascript\n" +
		"const greeting = (name) => {\n" +
		"    console.log(`Hello, ${name}!`);\n" +
		"};\n" +
		"```\n\n" +
		"### Blockquotes\n\n" +
		"> This is a blockquote with **bold** text.\n" +
		"> You can use *inline formatting* inside quotes.\n\n" +
		"### Links and Images\n\n" +
		"Check out this [documentation](https://golang.org) for more info.\n\n" +
		"![Logo](https://example.com/logo.png)\n\n" +
		"---\n\n" +
		"## Conclusion\n\n" +
		"The **Markdown** function provides a *simple* and `colorful` way to display formatted text in the terminal!\n"

	// Display the markdown content
	display.Markdown(markdownContent)

	// Example with a simpler markdown
	display.NewLine(2)
	display.Header("Another Example")
	display.NewLine()

	simpleMarkdown := "## Quick Start\n\n" +
		"Follow these steps:\n\n" +
		"1. Install the package\n" +
		"2. Import it: `import \"github.com/snipwise/nova/nova-sdk/ui/display\"`\n" +
		"3. Call `display.Markdown(content)`\n\n" +
		"**That's it!** You're ready to go.\n\n" +
		"---\n\n" +
		"### Notes\n\n" +
		"- Works with standard Go libraries only\n" +
		"- No external dependencies\n" +
		"- Full color support\n"

	display.Markdown(simpleMarkdown)
}
