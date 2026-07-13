//go:build !js && !wasm

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

func init() {
	SetPasswordPrompt(nativePasswordPrompt)
}

// nativePasswordPrompt reads a password without echoing it to the terminal.
// Falls back to the shared buffered reader when it already has bytes
// buffered: term.ReadPassword reads straight off the fd, so it would never
// see input an earlier prompt (e.g. a preceding ResourcePrompt in the same
// flow) already read ahead into that buffer.
func nativePasswordPrompt(msg string, noColor bool) (string, error) {
	// \U0001f512 is the lock symbol 🔒
	if !noColor {
		fmt.Fprint(os.Stderr, color.New(color.FgHiRed, color.Bold).Sprint("\U0001f512 "+msg+" "))
	} else {
		fmt.Fprint(os.Stderr, "\U0001f512 "+msg+" ")
	}

	if stdinHasBuffered() {
		return readStdinLine()
	}

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after masked input
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(password)), nil
}
