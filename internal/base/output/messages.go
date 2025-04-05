package output

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// PrintSuccess prints a success message with green color and a checkmark
func PrintSuccess(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("✓ %s\n", msg)
	} else {
		fmt.Print(color.GreenString("✓ "))
		fmt.Println(msg)
	}
}

// FormatSuccess returns a formatted success string
func FormatSuccess(msg string, noColor bool) string {
	if noColor {
		return "successfully"
	}
	return color.GreenString("successfully")
}

// PrintResourceSuccess prints a success message for a resource operation
func PrintResourceSuccess(resourceType, action, uid string, noColor bool) {
	uidFormatted := FormatUID(uid, noColor)
	if strings.HasSuffix(action, "ed") {
		PrintSuccess("%s %s %s", noColor, resourceType, action, uidFormatted)
	} else {
		PrintSuccess("%s %sed %s", noColor, resourceType, action, uidFormatted)
	}
}

// PrintResourceCreated prints a standardized resource creation message
func PrintResourceCreated(resourceType, uid string, noColor bool) {
	PrintSuccess("%s created %s", noColor, resourceType, FormatUID(uid, noColor))
}

// PrintResourceUpdated prints a standardized resource update message
func PrintResourceUpdated(resourceType, uid string, noColor bool) {
	PrintSuccess("%s updated %s", noColor, resourceType, FormatUID(uid, noColor))
}

// PrintResourceDeleted prints a standardized resource deletion message
func PrintResourceDeleted(resourceType, uid string, immediate, noColor bool) {
	msg := fmt.Sprintf("%s deleted %s", resourceType, FormatUID(uid, noColor))
	if immediate {
		msg += "\nThe resource will be deleted immediately"
	} else {
		msg += "\nThe resource will be deleted at the end of the current billing period"
	}
	PrintSuccess(msg, noColor)
}

// PrintError prints an error message with red color
func PrintError(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("✗ %s\n", msg)
	} else {
		fmt.Print(color.RedString("✗ "))
		fmt.Println(msg)
	}
}

// PrintWarning prints a warning message with yellow color
func PrintWarning(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("⚠ %s\n", msg)
	} else {
		fmt.Print(color.YellowString("⚠ "))
		fmt.Println(msg)
	}
}

// PrintInfo prints an info message with blue color
func PrintInfo(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("ℹ %s\n", msg)
	} else {
		fmt.Print(color.BlueString("ℹ "))
		fmt.Println(msg)
	}
}

// FormatConfirmation returns a formatted confirmation prompt
func FormatConfirmation(msg string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("%s [y/N]", msg)
	}
	return fmt.Sprintf("%s %s", msg, color.YellowString("[y/N]"))
}

// FormatPrompt returns a formatted input prompt
func FormatPrompt(msg string, noColor bool) string {
	if noColor {
		return msg
	}
	return color.BlueString(msg)
}

// FormatExample colorizes command examples in help text
func FormatExample(example string, noColor bool) string {
	if noColor {
		return example
	}
	return color.CyanString(example)
}

// FormatCommandName colorizes command names
func FormatCommandName(name string, noColor bool) string {
	if noColor {
		return name
	}
	return color.MagentaString(name)
}

// FormatRequiredFlag highlights required flags in help text
func FormatRequiredFlag(flag string, description string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("%s (REQUIRED): %s", flag, description)
	}
	return fmt.Sprintf("%s: %s", color.YellowString("%s (REQUIRED)", flag), description)
}

// FormatOptionalFlag formats optional flags in help text
func FormatOptionalFlag(flag string, description string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("%s: %s", flag, description)
	}
	return fmt.Sprintf("%s: %s", color.BlueString(flag), description)
}

// FormatJSONExample formats JSON examples in help text
func FormatJSONExample(json string, noColor bool) string {
	if noColor {
		return json
	}
	return color.GreenString(json)
}

func colorizeStatus(status string, noColor bool) string {
	if noColor {
		return status
	}

	upperStatus := strings.ToUpper(status)

	// Green for ready/active states
	switch upperStatus {
	case "CONFIGURED", "LIVE", "ACTIVE", "SUCCESS", "NEW":
		return color.GreenString(status)

	// Yellow for in-progress states
	case "CONFIGURING", "PROVISIONING", "PENDING", "REQUESTED", "DEPLOYING", "DEPLOYMENT":
		return color.YellowString(status)

	// Red for error/terminated states
	case "DECOMMISSIONED", "CANCELLED", "ERROR", "FAILED", "INACTIVE", "REJECTED", "RESTRICTED":
		return color.RedString(status)

	// Blue for informational states
	case "LOCKED", "MAINTENANCE", "SUSPENDED":
		return color.BlueString(status)

	// Default with no coloring
	default:
		return status
	}
}

func FormatUID(uid string, noColor bool) string {
	if noColor {
		return uid
	}
	return color.CyanString(uid)
}

// stripANSIColors removes ANSI color codes from a string
func StripANSIColors(s string) string {
	re := regexp.MustCompile("\x1b\\[[0-9;]*m")
	return re.ReplaceAllString(s, "")
}
