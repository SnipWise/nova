package main

import (
	"fmt"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	fmt.Println("=== Complete Test - No Duplicates + Code Blocks ===\n")

	parser := display.NewMarkdownChunkParser()

	// Complete test with various markdown elements
	testContent := `# HTTP Server in Go

Here's how to create a simple HTTP server in **Go**:

## Code Example

` + "```go\n" + `package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
` + "```\n" + `

## Explanation

This server does the following:

1. **Imports** the necessary packages
2. **Defines a handler** function that responds with "Hello, World!"
3. **Registers** the handler for the root path
4. **Starts** the server on port 8080

> Remember to test it with: curl http://localhost:8080
`

	// Simulate streaming by sending the content in chunks
	chunkSize := 50
	for i := 0; i < len(testContent); i += chunkSize {
		end := i + chunkSize
		if end > len(testContent) {
			end = len(testContent)
		}
		display.MarkdownChunk(parser, testContent[i:end])
	}
	parser.Flush()

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("✓ Code should be visible")
	fmt.Println("✓ No duplicates should appear")
}
