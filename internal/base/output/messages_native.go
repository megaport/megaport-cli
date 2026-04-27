//go:build !wasm
// +build !wasm

package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/megaport/megaport-cli/internal/base/exitcodes"
)

// PrintErrorJSON writes a structured JSON error to stderr.
// Used by RunE wrappers when --output json is active so automation scripts
// can parse errors programmatically instead of scraping plain text.
func PrintErrorJSON(code int, message string) {
	payload := errorEnvelope{
		Error: errorBody{
			Code:    code,
			Type:    exitcodes.TypeName(code),
			Message: message,
		},
	}
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload) // best-effort; stderr write failures are not actionable
}

// Native (non-WASM) implementations that write to stdout/stderr directly

func PrintSuccess(format string, noColor bool, args ...interface{}) {
	if IsQuiet() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if GetOutputFormat() == "json" {
		if noColor {
			fmt.Fprintf(os.Stderr, "✓ %s\n", msg)
		} else {
			fmt.Fprint(os.Stderr, color.GreenString("✓ "))
			fmt.Fprintln(os.Stderr, msg)
		}
	} else {
		if noColor {
			fmt.Printf("✓ %s\n", msg)
		} else {
			fmt.Print(color.GreenString("✓ "))
			fmt.Println(msg)
		}
	}
}

func PrintError(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if GetOutputFormat() == "json" {
		if noColor {
			fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
		} else {
			fmt.Fprint(os.Stderr, color.RedString("✗ "))
			fmt.Fprintln(os.Stderr, msg)
		}
	} else {
		if noColor {
			fmt.Printf("✗ %s\n", msg)
		} else {
			fmt.Print(color.RedString("✗ "))
			fmt.Println(msg)
		}
	}
}

func PrintWarning(format string, noColor bool, args ...interface{}) {
	if IsQuiet() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if GetOutputFormat() == "json" {
		if noColor {
			fmt.Fprintf(os.Stderr, "⚠ %s\n", msg)
		} else {
			fmt.Fprint(os.Stderr, color.YellowString("⚠ "))
			fmt.Fprintln(os.Stderr, msg)
		}
	} else {
		if noColor {
			fmt.Printf("⚠ %s\n", msg)
		} else {
			fmt.Print(color.YellowString("⚠ "))
			fmt.Println(msg)
		}
	}
}

func PrintInfo(format string, noColor bool, args ...interface{}) {
	if IsQuiet() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if GetOutputFormat() == "json" {
		if noColor {
			fmt.Fprintf(os.Stderr, "ℹ %s\n", msg)
		} else {
			fmt.Fprint(os.Stderr, color.BlueString("ℹ "))
			fmt.Fprintln(os.Stderr, msg)
		}
	} else {
		if noColor {
			fmt.Printf("ℹ %s\n", msg)
		} else {
			fmt.Print(color.BlueString("ℹ "))
			fmt.Println(msg)
		}
	}
}

// ClearScreen clears the terminal screen using ANSI escape codes.
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}
