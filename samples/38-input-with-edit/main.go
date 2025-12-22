package main

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	fmt.Println("=== Interactive Input Editing Example ===")

	// Example 1: Simple input with editing
	fmt.Println("1. Simple input with arrow key support:")
	input1 := prompt.NewWithColor("🤖 Ask me something?")
	question1, err := input1.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("You asked: %s\n\n", question1)

	// Example 2: Input with default value
	fmt.Println("2. Input with default value (try arrows to edit):")
	input2 := prompt.NewWithColor("📝 What is your name?").
		SetDefault("John Doe")
	name, err := input2.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello, %s!\n\n", name)

	// Example 3: Input with validation
	fmt.Println("3. Input with validation (must contain at least 5 characters):")
	input3 := prompt.NewWithColor("✍️  Enter a message").
		SetValidator(func(s string) error {
			if len(s) < 5 {
				return fmt.Errorf("message must contain at least 5 characters")
			}
			return nil
		})
	message, err := input3.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Validated message: %s\n\n", message)

	// Example 4: Input with custom colors
	fmt.Println("4. Input with custom colors:")
	input4 := prompt.NewWithColor("🎨 What is your favorite color?").
		SetMessageColor(prompt.ColorBrightMagenta).
		SetInputColor(prompt.ColorBrightCyan).
		SetDefault("blue")
	color, err := input4.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Your favorite color is: %s\n\n", color)

	// Example 5: Blinking cursor
	fmt.Println("5. Input with blinking cursor (block):")
	input5 := prompt.NewWithColor("⚡ Question with blinking cursor").
		SetCursorStyle(prompt.CursorBlockBlink).
		SetDefault("The cursor blinks!")
	result5, err := input5.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("You typed: %s\n\n", result5)

	// Example 6: Underline cursor
	fmt.Println("6. Input with underline cursor:")
	input6 := prompt.NewWithColor("📏 Question with underline cursor").
		SetCursorStyle(prompt.CursorUnderline).
		SetDefault("Cursor in underline mode")
	result6, err := input6.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("You typed: %s\n\n", result6)

	// Example 7: Blinking underline cursor
	fmt.Println("7. Input with blinking underline cursor:")
	input7 := prompt.NewWithColor("✨ Question with blinking underline cursor").
		SetCursorStyle(prompt.CursorUnderlineBlink).
		SetDefault("Cursor underline + blink")
	result7, err := input7.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("You typed: %s\n\n", result7)

	// Example 8: Keyboard shortcuts demonstration
	fmt.Println("8. Try the following keyboard shortcuts:")
	fmt.Println("   - Left/Right arrows: move cursor")
	fmt.Println("   - Home (Ctrl+A): go to beginning")
	fmt.Println("   - End (Ctrl+E): go to end")
	fmt.Println("   - Backspace: delete previous character")
	fmt.Println("   - Delete: delete next character")
	fmt.Println("   - Ctrl+K: delete from cursor to end")
	fmt.Println("   - Ctrl+U: delete from beginning to cursor")
	fmt.Println("   - Ctrl+C: cancel")

	input8 := prompt.NewWithColor("⌨️  Test the shortcuts").
		SetDefault("Edit this text with arrows!")
	result, err := input8.RunWithEdit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Final result: %s\n", result)
}
