package display

import (
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/ui/colors"
)

// Color codes for terminal output (re-exported from colors package for backward compatibility).
const (
	// Reset and modifiers
	ColorReset     = colors.ColorReset
	ColorBold      = colors.ColorBold
	ColorDim       = colors.ColorDim
	ColorItalic    = colors.ColorItalic
	ColorUnderline = colors.ColorUnderline
	ColorBlink     = colors.ColorBlink
	ColorReverse   = colors.ColorReverse
	ColorHidden    = colors.ColorHidden

	// Standard colors
	ColorBlack   = colors.ColorBlack
	ColorRed     = colors.ColorRed
	ColorGreen   = colors.ColorGreen
	ColorYellow  = colors.ColorYellow
	ColorBlue    = colors.ColorBlue
	ColorMagenta = colors.ColorMagenta
	ColorPurple  = colors.ColorPurple
	ColorCyan    = colors.ColorCyan
	ColorWhite   = colors.ColorWhite
	ColorGray    = colors.ColorGray

	// Bright colors
	ColorBrightBlack   = colors.ColorBrightBlack
	ColorBrightRed     = colors.ColorBrightRed
	ColorBrightGreen   = colors.ColorBrightGreen
	ColorBrightYellow  = colors.ColorBrightYellow
	ColorBrightBlue    = colors.ColorBrightBlue
	ColorBrightMagenta = colors.ColorBrightMagenta
	ColorBrightPurple  = colors.ColorBrightPurple
	ColorBrightCyan    = colors.ColorBrightCyan
	ColorBrightWhite   = colors.ColorBrightWhite

	// Background colors
	BgBlack   = colors.BgBlack
	BgRed     = colors.BgRed
	BgGreen   = colors.BgGreen
	BgYellow  = colors.BgYellow
	BgBlue    = colors.BgBlue
	BgMagenta = colors.BgMagenta
	BgCyan    = colors.BgCyan
	BgWhite   = colors.BgWhite

	// Bright background colors
	BgBrightBlack   = colors.BgBrightBlack
	BgBrightRed     = colors.BgBrightRed
	BgBrightGreen   = colors.BgBrightGreen
	BgBrightYellow  = colors.BgBrightYellow
	BgBrightBlue    = colors.BgBrightBlue
	BgBrightMagenta = colors.BgBrightMagenta
	BgBrightCyan    = colors.BgBrightCyan
	BgBrightWhite   = colors.BgBrightWhite
)

// Common symbols
const (
	SymbolSuccess = "✓"
	SymbolError   = "✗"
	SymbolWarning = "⚠"
	SymbolInfo    = "ℹ"
	SymbolDebug   = "●"
	SymbolArrow   = "→"
	SymbolBullet  = "•"
	SymbolCheck   = "✓"
	SymbolCross   = "✗"
	SymbolStar    = "★"
	SymbolHeart   = "♥"
	SymbolDiamond = "◆"
)

// Format strings for colored output
const (
	fmtColorLine     = "%s%s %s%s\n"  // color + symbol + text + reset (with space before text)
	fmtColorBlock    = "%s%s%s%s\n"   // color1 + color2 + text + reset
	fmtBlockHeaderNL = "\n%s%s%s%s\n" // leading newline + two colors + text + reset
)

// Message types
type MessageType int

const (
	MessagePlain MessageType = iota
	MessageSuccess
	MessageError
	MessageWarning
	MessageInfo
	MessageDebug
)

// Print prints a message without a newline
func Print(message string) {
	fmt.Print(message)
}

// Println prints a message with a newline
func Println(message string) {
	fmt.Println(message)
}

// Printf prints a formatted message
func Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

// Color prints a colored message without newline
func Color(message string, color string) {
	fmt.Printf("%s%s%s", color, message, ColorReset)
}

// Colorln prints a colored message with newline
func Colorln(message string, color string) {
	fmt.Printf("%s%s%s\n", color, message, ColorReset)
}

// Colorf prints a formatted colored message
func Colorf(color string, format string, args ...any) {
	fmt.Printf("%s%s%s", color, fmt.Sprintf(format, args...), ColorReset)
}

// Bold prints a bold message
func Bold(message string) {
	Color(message, ColorBold)
}

// Boldln prints a bold message with newline
func Boldln(message string) {
	Colorln(message, ColorBold)
}

// Italic prints an italic message
func Italic(message string) {
	Color(message, ColorItalic)
}

// Italicln prints an italic message with newline
func Italicln(message string) {
	Colorln(message, ColorItalic)
}

// Underline prints an underlined message
func Underline(message string) {
	Color(message, ColorUnderline)
}

// Underlineln prints an underlined message with newline
func Underlineln(message string) {
	Colorln(message, ColorUnderline)
}

// Success prints a success message with a checkmark
func Success(message string) {
	fmt.Printf(fmtColorLine, ColorGreen, SymbolSuccess, message, ColorReset)
}

// Successf prints a formatted success message
func Successf(format string, args ...any) {
	Success(fmt.Sprintf(format, args...))
}

// Error prints an error message with a cross
func Error(message string) {
	fmt.Printf(fmtColorLine, ColorRed, SymbolError, message, ColorReset)
}

// Errorf prints a formatted error message
func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}

// Warning prints a warning message with a warning symbol
func Warning(message string) {
	fmt.Printf(fmtColorLine, ColorYellow, SymbolWarning, message, ColorReset)
}

// Warningf prints a formatted warning message
func Warningf(format string, args ...any) {
	Warning(fmt.Sprintf(format, args...))
}

// Info prints an info message with an info symbol
func Info(message string) {
	fmt.Printf(fmtColorLine, ColorCyan, SymbolInfo, message, ColorReset)
}

// Infof prints a formatted info message
func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

// Debug prints a debug message
func Debug(message string) {
	fmt.Printf(fmtColorLine, ColorGray, SymbolDebug, message, ColorReset)
}

// Debugf prints a formatted debug message
func Debugf(format string, args ...any) {
	Debug(fmt.Sprintf(format, args...))
}

// Header prints a header message (bold and colored)
func Header(message string) {
	fmt.Printf(fmtColorBlock, ColorBold, ColorBrightCyan, message, ColorReset)
}

// Headerf prints a formatted header message
func Headerf(format string, args ...any) {
	Header(fmt.Sprintf(format, args...))
}

// Subheader prints a subheader message
func Subheader(message string) {
	fmt.Printf("%s%s%s\n", ColorBrightBlue, message, ColorReset)
}

// Subheaderf prints a formatted subheader message
func Subheaderf(format string, args ...any) {
	Subheader(fmt.Sprintf(format, args...))
}

// Title prints a title with a separator line
func Title(message string) {
	fmt.Printf(fmtBlockHeaderNL, ColorBold, ColorBrightCyan, message, ColorReset)
	fmt.Println(strings.Repeat("─", len(message)))
}

// Titlef prints a formatted title
func Titlef(format string, args ...any) {
	Title(fmt.Sprintf(format, args...))
}

// Separator prints a separator line
func Separator() {
	fmt.Println(strings.Repeat("─", 80))
}

// SeparatorWithChar prints a separator with a custom character
func SeparatorWithChar(char string, length int) {
	fmt.Println(strings.Repeat(char, length))
}

// Bullet prints a bulleted item
func Bullet(message string) {
	fmt.Printf("%s %s\n", SymbolBullet, message)
}

// Bulletf prints a formatted bulleted item
func Bulletf(format string, args ...any) {
	Bullet(fmt.Sprintf(format, args...))
}

// ColoredBullet prints a colored bulleted item
func ColoredBullet(message string, color string) {
	fmt.Printf(fmtColorLine, color, SymbolBullet, message, ColorReset)
}

// Arrow prints an arrow-prefixed message
func Arrow(message string) {
	fmt.Printf(fmtColorLine, ColorCyan, SymbolArrow, message, ColorReset)
}

// Arrowf prints a formatted arrow-prefixed message
func Arrowf(format string, args ...any) {
	Arrow(fmt.Sprintf(format, args...))
}

// Highlight prints a highlighted message (with background color)
func Highlight(message string, fgColor, bgColor string) {
	fmt.Printf(fmtColorBlock, bgColor, fgColor, message, ColorReset)
}

// Box prints a message in a box
func Box(message string) {
	width := len(message) + 4
	fmt.Printf("┌%s┐\n", strings.Repeat("─", width-2))
	fmt.Printf("│ %s │\n", message)
	fmt.Printf("└%s┘\n", strings.Repeat("─", width-2))
}

// ColoredBox prints a colored message in a box
func ColoredBox(message string, color string) {
	width := len(message) + 4
	fmt.Printf("%s┌%s┐\n", color, strings.Repeat("─", width-2))
	fmt.Printf("│ %s │\n", message)
	fmt.Printf("└%s┘%s\n", strings.Repeat("─", width-2), ColorReset)
}

// Banner prints a banner message
func Banner(message string) {
	width := len(message) + 4
	fmt.Printf("\n%s╔%s╗\n", ColorBold+ColorBrightYellow, strings.Repeat("═", width-2))
	fmt.Printf("║ %s%s%s%s ║\n", ColorBold, ColorBrightWhite, message, ColorBrightYellow)
	fmt.Printf("╚%s╝%s\n\n", strings.Repeat("═", width-2), ColorReset)
}

// List prints a numbered list item
func List(index int, message string) {
	fmt.Printf("%s%d.%s %s\n", ColorGray, index, ColorReset, message)
}

// Listf prints a formatted numbered list item
func Listf(index int, format string, args ...any) {
	List(index, fmt.Sprintf(format, args...))
}

// ColoredList prints a colored numbered list item
func ColoredList(index int, message string, color string) {
	fmt.Printf("%s%d. %s%s\n", color, index, message, ColorReset)
}

// Step prints a step indicator (for multi-step processes)
func Step(current, total int, message string) {
	fmt.Printf("%s[%d/%d]%s %s%s%s\n",
		ColorGray, current, total, ColorReset,
		ColorBrightCyan, message, ColorReset)
}

// Stepf prints a formatted step indicator
func Stepf(current, total int, format string, args ...any) {
	Step(current, total, fmt.Sprintf(format, args...))
}

// Progress prints a progress message
func Progress(message string) {
	fmt.Printf("%s⏳ %s...%s\n", ColorYellow, message, ColorReset)
}

// Progressf prints a formatted progress message
func Progressf(format string, args ...any) {
	Progress(fmt.Sprintf(format, args...))
}

// Done prints a completion message
func Done(message string) {
	fmt.Printf("%s✓ %s%s\n", ColorGreen, message, ColorReset)
}

// Donef prints a formatted completion message
func Donef(format string, args ...any) {
	Done(fmt.Sprintf(format, args...))
}

// NewLine prints one or more newlines
func NewLine(count ...int) {
	n := 1
	if len(count) > 0 {
		n = count[0]
	}
	for i := 0; i < n; i++ {
		fmt.Println()
	}
}

// Clear clears a line (useful for replacing spinner output)
func Clear() {
	fmt.Print("\r\033[K")
}

// Styled prints a message with custom styling
func Styled(message string, styles ...string) {
	fmt.Print(strings.Join(styles, ""))
	fmt.Print(message)
	fmt.Print(ColorReset)
}

// Styledln prints a styled message with newline
func Styledln(message string, styles ...string) {
	Styled(message, styles...)
	fmt.Println()
}

// Styledf prints a formatted styled message
func Styledf(format string, styles []string, args ...any) {
	Styled(fmt.Sprintf(format, args...), styles...)
}

// Indent prints an indented message
func Indent(level int, message string) {
	fmt.Printf("%s%s\n", strings.Repeat("  ", level), message)
}

// Indentf prints a formatted indented message
func Indentf(level int, format string, args ...any) {
	Indent(level, fmt.Sprintf(format, args...))
}

// ColoredIndent prints a colored indented message
func ColoredIndent(level int, message string, color string) {
	fmt.Printf(fmtColorBlock, strings.Repeat("  ", level), color, message, ColorReset)
}

// Table prints a simple 2-column table row
func Table(key, value string) {
	fmt.Printf("  %s%-20s%s %s%s%s\n",
		ColorGray, key+":", ColorReset,
		ColorWhite, value, ColorReset)
}

// Tablef prints a formatted 2-column table row
func Tablef(key string, format string, args ...any) {
	Table(key, fmt.Sprintf(format, args...))
}

// KeyValue prints a key-value pair
func KeyValue(key, value string) {
	fmt.Printf("%s%s:%s %s\n", ColorCyan, key, ColorReset, value)
}

// KeyValuef prints a formatted key-value pair
func KeyValuef(key, format string, args ...any) {
	KeyValue(key, fmt.Sprintf(format, args...))
}

// JSON-like output helpers

// ObjectStart prints the start of an object-like structure
func ObjectStart(name string) {
	fmt.Printf("%s%s {%s\n", ColorBrightCyan, name, ColorReset)
}

// ObjectEnd prints the end of an object-like structure
func ObjectEnd() {
	fmt.Println("}")
}

// Field prints a field in an object-like structure
func Field(key, value string) {
	fmt.Printf("  %s%s:%s %s\n", ColorCyan, key, ColorReset, value)
}

// Fieldf prints a formatted field
func Fieldf(key, format string, args ...any) {
	Field(key, fmt.Sprintf(format, args...))
}
