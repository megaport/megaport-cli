//go:build !wasm

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

// All status messages go to stderr so stdout carries only formatted data.

func PrintSuccess(format string, noColor bool, args ...interface{}) {
	if IsQuiet() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Fprintf(os.Stderr, "✓ %s\n", msg)
	} else {
		fmt.Fprint(os.Stderr, color.GreenString("✓ "))
		fmt.Fprintln(os.Stderr, msg)
	}
}

func PrintError(format string, noColor bool, args ...interface{}) {
	markErrorPrinted()
	// In JSON mode the structured envelope (PrintErrorJSON) is the single error
	// representation, so the human line is suppressed to avoid double output.
	if GetOutputFormat() == "json" {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	} else {
		fmt.Fprint(os.Stderr, color.RedString("✗ "))
		fmt.Fprintln(os.Stderr, msg)
	}
}

// PrintErrorPlain writes a human-readable error line to stderr unconditionally,
// regardless of output format. The RunE wrappers use it to surface a failure
// the action did not print itself; it does not mark the printed-error flag.
func PrintErrorPlain(message string, noColor bool) {
	if noColor {
		fmt.Fprintf(os.Stderr, "✗ %s\n", message)
	} else {
		fmt.Fprint(os.Stderr, color.RedString("✗ "))
		fmt.Fprintln(os.Stderr, message)
	}
}

func PrintWarning(format string, noColor bool, args ...interface{}) {
	if IsQuiet() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", msg)
	} else {
		fmt.Fprint(os.Stderr, color.YellowString("⚠ "))
		fmt.Fprintln(os.Stderr, msg)
	}
}

func PrintInfo(format string, noColor bool, args ...interface{}) {
	if IsQuiet() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Fprintf(os.Stderr, "ℹ %s\n", msg)
	} else {
		fmt.Fprint(os.Stderr, color.BlueString("ℹ "))
		fmt.Fprintln(os.Stderr, msg)
	}
}

// PrintPlain prints a plain line with no icon prefix. Like PrintInfo it is
// suppressed in quiet mode and routed to stderr in JSON output mode.
func PrintPlain(format string, _ bool, args ...interface{}) {
	if IsQuiet() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if GetOutputFormat() == "json" {
		fmt.Fprintln(os.Stderr, msg)
	} else {
		fmt.Println(msg)
	}
}

// PrintNewline prints a blank line, suppressed in quiet mode and routed to
// stderr when the output format is json (consistent with PrintInfo et al.).
func PrintNewline() {
	if IsQuiet() {
		return
	}
	if GetOutputFormat() == "json" {
		fmt.Fprintln(os.Stderr)
		return
	}
	fmt.Println()
}

// ClearScreen clears the terminal screen using ANSI escape codes.
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}
