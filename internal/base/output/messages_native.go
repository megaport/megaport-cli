//go:build !wasm
// +build !wasm

package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Native (non-WASM) implementations that write to stdout/stderr directly

func PrintSuccess(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if currentOutputFormat == "json" {
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
	if currentOutputFormat == "json" {
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
	msg := fmt.Sprintf(format, args...)
	if currentOutputFormat == "json" {
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
	msg := fmt.Sprintf(format, args...)
	if currentOutputFormat == "json" {
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
