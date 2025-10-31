package output

import (
	"strings"

	"github.com/fatih/color"
)

var (
	RadRed           = color.New(color.FgRed).Add(color.Bold).SprintFunc()
	DeepNightBlue    = color.New(color.FgBlue).Add(color.FgBlack).SprintFunc()
	MegaportYellow   = color.New(color.FgHiYellow).SprintFunc()
	MegaportOrange   = color.New(color.FgHiRed).SprintFunc()
	MegaportRed      = color.New(color.FgRed).SprintFunc()
	MegaportPurple   = color.New(color.FgMagenta).SprintFunc()
	RadRedReadable   = color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	DeepBlueReadable = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	YellowReadable   = color.New(color.FgYellow).Add(color.Bold).SprintFunc()
	OrangeReadable   = color.New(color.FgRed).Add(color.FgYellow).SprintFunc()
	PurpleReadable   = color.New(color.FgHiMagenta).Add(color.Bold).SprintFunc()
	NexusBlue        = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	DodgerBlue       = color.New(color.FgHiBlue).Add(color.Bold).SprintFunc()
	SkyBlue          = color.New(color.FgHiCyan).SprintFunc()
	PurpleCloud      = color.New(color.FgMagenta).Add(color.Bold).SprintFunc()
	Plum             = color.New(color.FgHiMagenta).SprintFunc()
	Mauve            = color.New(color.FgHiMagenta).Add(color.Faint).SprintFunc()
	Pink             = color.New(color.FgMagenta).Add(color.Italic).SprintFunc()
	Magenta          = color.New(color.FgHiMagenta).Add(color.Bold).SprintFunc()
	SunsetOrange     = color.New(color.FgRed).Add(color.FgHiYellow).SprintFunc()
	BlackHole        = color.New(color.FgBlack).Add(color.Bold).SprintFunc()
	DarkBlue         = color.New(color.FgBlue).Add(color.Underline).SprintFunc()
	GoldYellow       = color.New(color.FgYellow).Add(color.Bold).SprintFunc()
	Teal             = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
	LinkGreen        = color.New(color.FgGreen).Add(color.Bold).SprintFunc()
	ElectricLime     = color.New(color.FgHiGreen).SprintFunc()
)

func colorizeValue(val string, header string, noColor bool) string {
	if noColor {
		return val
	}
	if header == "status" || header == "provisioning_status" || strings.Contains(header, "state") {
		return colorizeStatus(val, noColor)
	} else if strings.HasSuffix(header, "uid") || strings.HasSuffix(header, "id") {
		return DodgerBlue(val)
	} else if strings.Contains(header, "price") || strings.Contains(header, "cost") ||
		strings.Contains(header, "rate") {
		return GoldYellow(val)
	} else if header == "name" || header == "product_name" || header == "title" {
		return RadRed(val)
	} else if strings.Contains(header, "speed") || strings.Contains(header, "bandwidth") {
		return DodgerBlue(val)
	} else if header == "location_id" || header == "locationid" || header == "metro" || header == "country" {
		return Teal(val)
	} else if strings.Contains(header, "ip") || strings.Contains(header, "cidr") || strings.Contains(header, "subnet") {
		return LinkGreen(val)
	} else if strings.Contains(header, "vlan") || strings.Contains(header, "asn") {
		return SkyBlue(val)
	} else if strings.Contains(header, "type") || strings.Contains(header, "product") {
		return RadRed(val)
	} else if strings.Contains(header, "port") || strings.Contains(header, "interface") {
		return PurpleCloud(val)
	} else if val == "" || val == "<nil>" || val == "null" {
		return color.New(color.FgHiBlack).Sprint("<empty>")
	} else if val == "true" {
		return LinkGreen(val)
	} else if val == "false" {
		return SunsetOrange(val)
	}
	return color.New(color.FgWhite).Add(color.Bold).Sprint(val)
}

func colorizeStatus(status string, noColor bool) string {
	if noColor {
		return status
	}
	status = strings.ToUpper(status)
	
	// Create bordered badges for better visibility
	// Format: [ STATUS ] with colored background and borders
	
	if strings.Contains(status, "ACTIVE") || strings.Contains(status, "LIVE") ||
		strings.Contains(status, "CONFIGURED") || status == "UP" || status == "AVAILABLE" {
		// Green badge with bright background
		return color.New(color.FgHiWhite, color.BgGreen, color.Bold).Sprintf(" %s ", status)
	}
	if strings.Contains(status, "PENDING") || strings.Contains(status, "PROVISIONING") ||
		strings.Contains(status, "WAITING") || strings.Contains(status, "REQUESTED") ||
		strings.Contains(status, "DEPLOYABLE") {
		// Yellow badge for in-progress states
		return color.New(color.FgBlack, color.BgYellow, color.Bold).Sprintf(" %s ", status)
	}
	if strings.Contains(status, "ERROR") || strings.Contains(status, "FAILED") {
		// Red badge for error states
		return color.New(color.FgHiWhite, color.BgRed, color.Bold).Sprintf(" %s ", status)
	}
	if strings.Contains(status, "CANCELLED") || strings.Contains(status, "DELETED") ||
		status == "DOWN" || strings.Contains(status, "INACTIVE") ||
		strings.Contains(status, "DECOMMISSIONING") || strings.Contains(status, "DECOMMISSIONED") {
		// Orange badge for terminated states
		return color.New(color.FgHiWhite, color.BgHiRed, color.Bold).Sprintf(" %s ", status)
	}
	if strings.Contains(status, "DEGRADED") || strings.Contains(status, "PARTIAL") {
		// Magenta badge for degraded states
		return color.New(color.FgHiWhite, color.BgMagenta, color.Bold).Sprintf(" %s ", status)
	}
	if strings.Contains(status, "RESERVED") || strings.Contains(status, "LOCKED") {
		// Purple badge for locked states
		return color.New(color.FgHiWhite, color.BgHiMagenta, color.Bold).Sprintf(" %s ", status)
	}
	// Blue badge for other states
	return color.New(color.FgHiWhite, color.BgBlue, color.Bold).Sprintf(" %s ", status)
}
