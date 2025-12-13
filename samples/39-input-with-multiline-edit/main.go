package main

import (
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	fmt.Println("=== Multi-Line Input Editor Example ===")
	fmt.Println()

	// Example 1: Simple multi-line input
	fmt.Println("1. Simple multi-line input:")
	fmt.Println("   - Press Enter to create new lines")
	fmt.Println("   - Press Ctrl+D to submit")
	fmt.Println()
	input1 := prompt.NewWithColor("📝 Write a short poem")
	poem, err := input1.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nYour poem:\n%s\n\n", poem)
	fmt.Println(strings.Repeat("=", 50))

	// Example 2: Multi-line input with default value
	fmt.Println("\n2. Multi-line input with default value (try editing):")
	defaultText := `Line 1: Hello World`
	input2 := prompt.NewWithColor("✏️  Edit this text").
		SetDefault(defaultText)
	edited, err := input2.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nEdited text:\n%s\n\n", edited)
	fmt.Println(strings.Repeat("=", 50))

	// Example 3: Multi-line input with validation
	fmt.Println("\n3. Multi-line input with validation (must have at least 3 lines):")
	input3 := prompt.NewWithColor("📋 Enter a list (minimum 3 lines)").
		SetValidator(func(s string) error {
			lines := strings.Split(s, "\n")
			if len(lines) < 3 {
				return fmt.Errorf("you must enter at least 3 lines (current: %d)", len(lines))
			}
			return nil
		})
	list, err := input3.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nYour list (%d lines):\n%s\n\n", len(strings.Split(list, "\n")), list)
	fmt.Println(strings.Repeat("=", 50))

	// Example 4: Code snippet input
	fmt.Println("\n4. Enter a code snippet:")
	input4 := prompt.NewWithColor("💻 Paste or write your code").
		SetMessageColor(prompt.ColorBrightCyan).
		SetInputColor(prompt.ColorBrightYellow)
	code, err := input4.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nYour code:\n%s\n\n", code)
	fmt.Println(strings.Repeat("=", 50))

	// Example 5: Multi-line with blinking cursor
	fmt.Println("\n5. Multi-line input with blinking cursor:")
	input5 := prompt.NewWithColor("⚡ Write something with a blinking cursor").
		SetCursorStyle(prompt.CursorBlockBlink)
	result5, err := input5.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nYou wrote:\n%s\n\n", result5)
	fmt.Println(strings.Repeat("=", 50))

	// Example 6: Multi-line with underline cursor
	fmt.Println("\n6. Multi-line input with underline cursor:")
	input6 := prompt.NewWithColor("📏 Try the underline cursor style").
		SetCursorStyle(prompt.CursorUnderline).
		SetDefault("Line 1\nLine 2\nLine 3")
	result6, err := input6.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nResult:\n%s\n\n", result6)
	fmt.Println(strings.Repeat("=", 50))

	// Example 7: Keyboard shortcuts demonstration
	fmt.Println("\n7. Multi-line keyboard shortcuts reference:")
	fmt.Println("   Navigation:")
	fmt.Println("   - Up/Down arrows (Ctrl+P/N): move between lines")
	fmt.Println("   - Left/Right arrows (Ctrl+B/F): move within line")
	fmt.Println("   - Home (Ctrl+A): go to line beginning")
	fmt.Println("   - End (Ctrl+E): go to line end")
	fmt.Println()
	fmt.Println("   Editing:")
	fmt.Println("   - Enter: insert new line")
	fmt.Println("   - Backspace: delete previous character (or merge lines)")
	fmt.Println("   - Delete: delete next character")
	fmt.Println("   - Ctrl+K: delete from cursor to end of line")
	fmt.Println("   - Ctrl+U: delete from beginning of line to cursor")
	fmt.Println()
	fmt.Println("   Control:")
	fmt.Println("   - Ctrl+D: submit the input")
	fmt.Println("   - Ctrl+C: cancel (exit program)")
	fmt.Println()

	input7 := prompt.NewWithColor("⌨️  Try all the shortcuts").
		SetDefault("Line 1: Navigate with arrows\nLine 2: Edit this text\nLine 3: Press Ctrl+D when done")
	result7, err := input7.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nFinal result:\n%s\n\n", result7)
	fmt.Println(strings.Repeat("=", 50))

	// Example 8: Writing a message/email
	fmt.Println("\n8. Compose a message:")
	input8 := prompt.NewWithColor("✉️  Write your message").
		SetMessageColor(prompt.ColorBrightGreen).
		SetInputColor(prompt.ColorBrightWhite).
		SetValidator(func(s string) error {
			if len(strings.TrimSpace(s)) == 0 {
				return fmt.Errorf("message cannot be empty")
			}
			return nil
		})
	message, err := input8.RunWithMultiLineEdit()
	if err != nil {
		panic(err)
	}

	lines := strings.Split(message, "\n")
	words := len(strings.Fields(message))
	chars := len(message)

	fmt.Printf("\nYour message:\n%s\n\n", message)
	fmt.Printf("Statistics: %d lines, %d words, %d characters\n", len(lines), words, chars)
	fmt.Println(strings.Repeat("=", 50))

	fmt.Println("\n✅ All examples completed!")
}
