package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// prompt is a helper function to prompt the user for input
func prompt(message string) (string, error) {
	fmt.Print(message)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
