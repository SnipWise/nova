package main

import (
	"fmt"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	fmt.Println("=== Testing Markdown Streaming Fix ===\n")

	parser := display.NewMarkdownChunkParser()

	// Simulate streaming content that arrives in chunks
	testChunks := []string{
		"This is ",
		"**bold** ",
		"and ",
		"*italic",
		"* text.\n",
		"## Header\n",
		"- Item ",
		"1\n",
		"- Item 2\n",
	}

	for _, chunk := range testChunks {
		display.MarkdownChunk(parser, chunk)
	}
	parser.Flush()

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("There should be NO duplicates above!")
}
