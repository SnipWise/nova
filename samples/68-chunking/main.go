package main

import "github.com/snipwise/nova/nova-sdk/agents/rag/chunks"

func main() {
	doc := `
	# My Document

	## My Life

	### Childhood
	I was born in a small town...

	### Adulthood

	I moved to the city and started my career...
	
	## My Work
	I work as a software developer...
	`

	chunks := chunks.SplitMarkdownBySection(2,doc)

	for i, chunk := range chunks {
		println("---- Chunk", i, "----")
		println(chunk)
	}
}


