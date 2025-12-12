package main

import (
	"fmt"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	fmt.Println("=== Testing Code Block Rendering ===\n")

	parser := display.NewMarkdownChunkParser()

	// Simulate streaming a code block
	testChunks := []string{
		"Here is some code:\n",
		"```go\n",
		"func main() {\n",
		"    fmt.Println(\"Hello, World!\")\n",
		"}\n",
		"```\n",
		"End of code.\n",
	}

	for _, chunk := range testChunks {
		display.MarkdownChunk(parser, chunk)
	}
	parser.Flush()

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("The code block should be visible above!")
}
