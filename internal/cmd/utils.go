package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var prompt = func(msg string) (string, error) {
	fmt.Print(msg)

	// Create a new reader for each prompt
	reader := bufio.NewReader(os.Stdin)

	// Read until newline and handle trimming properly
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Trim both spaces and newline characters from both ends
	return strings.TrimSpace(input), nil
}
