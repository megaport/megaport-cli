package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// PrintSuccess prints a success message with green color and a checkmark
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !noColor {
		fmt.Print(color.GreenString("✓ "))
		fmt.Println(msg)
	} else {
		fmt.Printf("✓ %s\n", msg)
	}
}

// FormatSuccess returns a formatted success string
func FormatSuccess(msg string) string {
	if !noColor {
		return color.GreenString("successfully")
	}
	return "successfully"
}

// PrintResourceSuccess prints a success message for a resource operation
func PrintResourceSuccess(resourceType, action, uid string) {
	uidFormatted := formatUID(uid)
	if strings.HasSuffix(action, "ed") {
		PrintSuccess("%s %s %s", resourceType, action, uidFormatted)
	} else {
		PrintSuccess("%s %sed %s", resourceType, action, uidFormatted)
	}
}

// PrintResourceCreated prints a standardized resource creation message
func PrintResourceCreated(resourceType, uid string) {
	PrintSuccess("%s created %s", resourceType, formatUID(uid))
}

// PrintResourceUpdated prints a standardized resource update message
func PrintResourceUpdated(resourceType, uid string) {
	PrintSuccess("%s updated %s", resourceType, formatUID(uid))
}

// PrintResourceDeleted prints a standardized resource deletion message
func PrintResourceDeleted(resourceType, uid string, immediate bool) {
	msg := fmt.Sprintf("%s deleted %s", resourceType, formatUID(uid))
	if immediate {
		msg += "\nThe resource will be deleted immediately"
	} else {
		msg += "\nThe resource will be deleted at the end of the current billing period"
	}
	PrintSuccess(msg)
}

// PrintError prints an error message with red color
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !noColor {
		fmt.Print(color.RedString("✗ "))
		fmt.Println(msg)
	} else {
		fmt.Printf("✗ %s\n", msg)
	}
}

// PrintWarning prints a warning message with yellow color
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !noColor {
		fmt.Print(color.YellowString("⚠ "))
		fmt.Println(msg)
	} else {
		fmt.Printf("⚠ %s\n", msg)
	}
}

// PrintInfo prints an info message with blue color
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !noColor {
		fmt.Print(color.BlueString("ℹ "))
		fmt.Println(msg)
	} else {
		fmt.Printf("ℹ %s\n", msg)
	}
}

// FormatConfirmation returns a formatted confirmation prompt
func FormatConfirmation(msg string) string {
	if !noColor {
		return fmt.Sprintf("%s %s", msg, color.YellowString("[y/N]"))
	}
	return fmt.Sprintf("%s [y/N]", msg)
}

// FormatPrompt returns a formatted input prompt
func FormatPrompt(msg string) string {
	if !noColor {
		return color.BlueString(msg)
	}
	return msg
}
