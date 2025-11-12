//go:build js && wasm
// +build js,wasm

package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/megaport/megaport-cli/internal/wasm"
)

// WASM-specific message formatting with prominent boxes and borders
// These functions enhance visibility in the browser terminal

// createBox creates a bordered box around text for prominent display
func createBox(text string, borderColor *color.Color, textColor *color.Color, width int) string {
	lines := strings.Split(text, "\n")
	
	// Calculate box width
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	if width > maxLen {
		maxLen = width
	}
	
	// Box drawing characters
	topLeft := "╔"
	topRight := "╗"
	bottomLeft := "╚"
	bottomRight := "╝"
	horizontal := "═"
	vertical := "║"
	
	// Build the box
	var result strings.Builder
	
	// Top border
	result.WriteString(borderColor.Sprint(topLeft))
	result.WriteString(borderColor.Sprint(strings.Repeat(horizontal, maxLen+2)))
	result.WriteString(borderColor.Sprint(topRight))
	result.WriteString("\n")
	
	// Content lines
	for _, line := range lines {
		padding := maxLen - len(line)
		result.WriteString(borderColor.Sprint(vertical))
		result.WriteString(" ")
		result.WriteString(textColor.Sprint(line))
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(" ")
		result.WriteString(borderColor.Sprint(vertical))
		result.WriteString("\n")
	}
	
	// Bottom border
	result.WriteString(borderColor.Sprint(bottomLeft))
	result.WriteString(borderColor.Sprint(strings.Repeat(horizontal, maxLen+2)))
	result.WriteString(borderColor.Sprint(bottomRight))
	
	return result.String()
}

// PrintSuccessBox prints a success message in a prominent green box
func PrintSuccessBox(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	
	if noColor {
		fmt.Printf("✓ %s\n", msg)
	} else {
		// Create a green bordered box for success messages
		borderColor := color.New(color.FgHiGreen, color.Bold)
		textColor := color.New(color.FgHiWhite, color.Bold)
		icon := color.New(color.FgHiGreen, color.Bold).Sprint("✓ ")
		box := createBox(icon+msg, borderColor, textColor, 40)
		fmt.Println(box)
	}
}

// PrintErrorBox prints an error message in a prominent red box
func PrintErrorBox(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	
	if noColor {
		fmt.Printf("✗ %s\n", msg)
	} else {
		// Create a red bordered box for error messages
		borderColor := color.New(color.FgHiRed, color.Bold)
		textColor := color.New(color.FgHiWhite, color.Bold)
		icon := color.New(color.FgHiRed, color.Bold).Sprint("✗ ")
		box := createBox(icon+msg, borderColor, textColor, 40)
		fmt.Println(box)
	}
}

// PrintWarningBox prints a warning message in a prominent yellow box
func PrintWarningBox(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	
	if noColor {
		fmt.Printf("⚠ %s\n", msg)
	} else {
		// Create a yellow bordered box for warning messages
		borderColor := color.New(color.FgHiYellow, color.Bold)
		textColor := color.New(color.FgBlack, color.Bold)
		icon := color.New(color.FgHiYellow, color.Bold).Sprint("⚠ ")
		box := createBox(icon+msg, borderColor, textColor, 40)
		fmt.Println(box)
	}
}

// PrintInfoBox prints an info message in a prominent blue box
func PrintInfoBox(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	
	if noColor {
		fmt.Printf("ℹ %s\n", msg)
	} else {
		// Create a blue bordered box for info messages
		borderColor := color.New(color.FgHiCyan, color.Bold)
		textColor := color.New(color.FgHiWhite, color.Bold)
		icon := color.New(color.FgHiCyan, color.Bold).Sprint("ℹ ")
		box := createBox(icon+msg, borderColor, textColor, 40)
		fmt.Println(box)
	}
}

// PrintBanner prints a large prominent banner message
func PrintBanner(text string, bannerType string, noColor bool) {
	if noColor {
		fmt.Println(text)
		return
	}
	
	var borderColor, textColor *color.Color
	
	switch bannerType {
	case "success":
		borderColor = color.New(color.FgHiGreen, color.Bold)
		textColor = color.New(color.FgHiWhite, color.BgGreen, color.Bold)
	case "error":
		borderColor = color.New(color.FgHiRed, color.Bold)
		textColor = color.New(color.FgHiWhite, color.BgRed, color.Bold)
	case "warning":
		borderColor = color.New(color.FgHiYellow, color.Bold)
		textColor = color.New(color.FgBlack, color.BgYellow, color.Bold)
	case "info":
		borderColor = color.New(color.FgHiCyan, color.Bold)
		textColor = color.New(color.FgHiWhite, color.BgCyan, color.Bold)
	default:
		borderColor = color.New(color.FgHiWhite, color.Bold)
		textColor = color.New(color.FgHiWhite, color.Bold)
	}
	
	// Create a wide banner
	banner := createBox(text, borderColor, textColor, 60)
	fmt.Println()
	fmt.Println(banner)
	fmt.Println()
}

// PrintProgressBox prints a progress message with a styled box
func PrintProgressBox(message string, percentage int, noColor bool) {
	if noColor {
		fmt.Printf("[%d%%] %s\n", percentage, message)
		return
	}
	
	// Create a progress bar within a box
	barWidth := 30
	filledWidth := (percentage * barWidth) / 100
	emptyWidth := barWidth - filledWidth
	
	filled := color.New(color.BgGreen).Sprint(strings.Repeat(" ", filledWidth))
	empty := color.New(color.BgHiBlack).Sprint(strings.Repeat(" ", emptyWidth))
	progressBar := fmt.Sprintf("%s%s", filled, empty)
	
	percentText := color.New(color.FgHiWhite, color.Bold).Sprintf("%3d%%", percentage)
	text := fmt.Sprintf("%s │%s│ %s", percentText, progressBar, message)
	
	borderColor := color.New(color.FgHiCyan, color.Bold)
	textColor := color.New(color.FgHiWhite)
	box := createBox(text, borderColor, textColor, 50)
	
	fmt.Print("\r") // Clear line
	fmt.Print(box)
}

// WASM-specific overrides for print functions to ensure output is captured
// These write to WasmOutputBuffer instead of stdout

// PrintSuccess overrides the base function for WASM to capture output
func PrintSuccess(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	var output string
	if noColor {
		output = fmt.Sprintf("✓ %s\n", msg)
	} else {
		output = color.GreenString("✓ ") + msg + "\n"
	}
	wasm.WasmOutputBuffer.Write([]byte(output))
}

// PrintError overrides the base function for WASM to capture output
func PrintError(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	var output string
	if noColor {
		output = fmt.Sprintf("✗ %s\n", msg)
	} else {
		output = color.RedString("✗ ") + msg + "\n"
	}
	wasm.WasmOutputBuffer.Write([]byte(output))
}

// PrintWarning overrides the base function for WASM to capture output
func PrintWarning(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	var output string
	if noColor {
		output = fmt.Sprintf("⚠ %s\n", msg)
	} else {
		output = color.YellowString("⚠ ") + msg + "\n"
	}
	wasm.WasmOutputBuffer.Write([]byte(output))
}

// PrintInfo overrides the base function for WASM to capture output
func PrintInfo(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	var output string
	if noColor {
		output = fmt.Sprintf("ℹ %s\n", msg)
	} else {
		output = color.BlueString("ℹ ") + msg + "\n"
	}
	wasm.WasmOutputBuffer.Write([]byte(output))
}
