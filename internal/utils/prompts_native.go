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

func nativePasswordPrompt(msg string, noColor bool) (string, error) {
	// \U0001f512 is the lock symbol 🔒
	if !noColor {
		fmt.Fprint(os.Stderr, color.New(color.FgHiRed, color.Bold).Sprint("\U0001f512 "+msg+" "))
	} else {
		fmt.Fprint(os.Stderr, "\U0001f512 "+msg+" ")
	}

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after masked input
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(password)), nil
}
