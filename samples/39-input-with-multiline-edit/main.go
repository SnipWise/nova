package main

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {

	input1 := prompt.NewWithColor("📝 Write a short poem")
	poem, err := input1.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nYour poem:\n%s\n\n", poem)

}
