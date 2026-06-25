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

// Native (non-WASM) implementations. All status and error messages are written
// to stderr regardless of the output format so that stdout carries only the
// formatted data stream (table/csv/xml/json).

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

// PrintError writes an error message to stderr and latches that an error was
// shown. In json mode it is a no-op: the RunE wrapper owns error output and
// emits a single structured envelope, so a human "✗" line would both duplicate
// it and corrupt the machine-readable stream.
func PrintError(format string, noColor bool, args ...interface{}) {
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
	markErrorEmitted()
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
