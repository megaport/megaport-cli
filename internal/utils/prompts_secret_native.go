//go:build !js || !wasm

package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

func init() {
	SetSecretResourcePrompt(nativeSecretResourcePrompt)
}

// nativeSecretResourcePrompt reads a sensitive value without echoing it to
// the terminal. Falls back to the standard echoed prompt when stdin is not a
// terminal (piped input, CI) so scripted usage keeps working.
func nativeSecretResourcePrompt(resourceType string, msg string, noColor bool) (string, error) {
	icon := "🔐"

	if !noColor {
		fmt.Print(color.New(color.FgHiRed, color.Bold).Sprint(icon + " " + msg + " "))
	} else {
		fmt.Print(icon + " " + msg + " ")
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(input), nil
	}

	pw, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(pw)), nil
}
