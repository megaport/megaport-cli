package cmd

import (
	"strings"

	"github.com/fatih/color"
)

// Helper File for Color
// This file contains functions to colorize text output in the terminal.

// Helper functions for standardized colored output

func colorizeStatus(status string) string {
	// Use no-color flag to disable colors if requested
	if noColor {
		return status
	}

	upperStatus := strings.ToUpper(status)

	// Green for ready/active states
	switch upperStatus {
	case "CONFIGURED", "LIVE", "ACTIVE", "SUCCESS", "NEW":
		// Aligned with SERVICE_CONFIGURED, SERVICE_LIVE, and SERVICE_STATE_READY
		return color.GreenString(status)

	// Yellow for in-progress states
	case "CONFIGURING", "PROVISIONING", "PENDING", "REQUESTED", "DEPLOYING", "DEPLOYMENT":
		return color.YellowString(status)

	// Red for error/terminated states
	case "DECOMMISSIONED", "CANCELLED", "ERROR", "FAILED", "INACTIVE", "REJECTED", "RESTRICTED":
		// Aligned with STATUS_DECOMMISSIONED, STATUS_CANCELLED
		return color.RedString(status)

	// Blue for informational states
	case "LOCKED", "MAINTENANCE", "SUSPENDED":
		return color.BlueString(status)

	// Default with no coloring
	default:
		return status
	}
}

func formatUID(uid string) string {
	if noColor {
		return uid
	}
	return color.CyanString(uid)
}

// func formatSectionHeader(title string) string {
// 	if noColor {
// 		return strings.ToUpper(title) + ":"
// 	}
// 	return color.New(color.Bold, color.FgHiBlue).Sprint(strings.ToUpper(title) + ":")
// }

// func formatDiff(oldVal, newVal string) string {
// 	if noColor {
// 		return fmt.Sprintf("%s → %s", oldVal, newVal)
// 	}
// 	return fmt.Sprintf("%s → %s",
// 		color.RedString(oldVal),
// 		color.GreenString(newVal))
// }
