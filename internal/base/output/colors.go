package output

import (
	"strings"

	"github.com/fatih/color"
)

// Megaport brand colors (from brand guidelines at https://www.megaport.com/branding/megaport-brand-guidelines.pdf)
var (
	// Primary colors
	RadRed        = color.New(color.FgRed).Add(color.Bold).SprintFunc()     // #E40046, PMS 192 C, RGB(228,0,70), CMYK(0,100,69,11)
	DeepNightBlue = color.New(color.FgBlue).Add(color.FgBlack).SprintFunc() // #0C1124, PMS 276 C, RGB(12,17,36), CMYK(67,53,0,86)

	// Secondary colors
	MegaportYellow = color.New(color.FgHiYellow).SprintFunc() // #FFBD39
	MegaportOrange = color.New(color.FgHiRed).SprintFunc()    // #FF4713
	MegaportRed    = color.New(color.FgRed).SprintFunc()      // #DC0032
	MegaportPurple = color.New(color.FgMagenta).SprintFunc()  // #764394

	// Readability-enhanced brand colors (modified for CLI readability while maintaining brand identity)
	RadRedReadable   = color.New(color.FgHiRed).Add(color.Bold).SprintFunc()     // Enhanced RadRed for better terminal contrast
	DeepBlueReadable = color.New(color.FgBlue).Add(color.Bold).SprintFunc()      // Enhanced DeepNightBlue for readability
	YellowReadable   = color.New(color.FgYellow).Add(color.Bold).SprintFunc()    // Enhanced MegaportYellow
	OrangeReadable   = color.New(color.FgRed).Add(color.FgYellow).SprintFunc()   // Enhanced MegaportOrange
	PurpleReadable   = color.New(color.FgHiMagenta).Add(color.Bold).SprintFunc() // Enhanced MegaportPurple

	// Accent colors
	NexusBlue    = color.New(color.FgBlue).Add(color.Bold).SprintFunc()       // #0072DA (PMS 285 C)
	DodgerBlue   = color.New(color.FgHiBlue).Add(color.Bold).SprintFunc()     // #1AA0FF (PMS 2925 C)
	SkyBlue      = color.New(color.FgHiCyan).SprintFunc()                     // #70D9F8 (PMS Blue 0821 C)
	PurpleCloud  = color.New(color.FgMagenta).Add(color.Bold).SprintFunc()    // #6500D1 (PMS 267 C)
	Plum         = color.New(color.FgHiMagenta).SprintFunc()                  // #A555F5 (PMS 265 C)
	Mauve        = color.New(color.FgHiMagenta).Add(color.Faint).SprintFunc() // #C49BF8 (PMS 264 C)
	Pink         = color.New(color.FgMagenta).Add(color.Italic).SprintFunc()  // #EA3388 (PMS 1915 C)
	Magenta      = color.New(color.FgHiMagenta).Add(color.Bold).SprintFunc()  // #E24BCC (PMS 238 C)
	SunsetOrange = color.New(color.FgRed).Add(color.FgHiYellow).SprintFunc()  // #FF7F32 (PMS 1575 C)
	BlackHole    = color.New(color.FgBlack).Add(color.Bold).SprintFunc()      // #000000 (PMS Black C)
	DarkBlue     = color.New(color.FgBlue).Add(color.Underline).SprintFunc()  // #200786 (PMS 2735 C)
	GoldYellow   = color.New(color.FgYellow).Add(color.Bold).SprintFunc()     // #FAAE3B (PMS 1365 C)
	Teal         = color.New(color.FgCyan).Add(color.Bold).SprintFunc()       // #00ACB6 (PMS 7466 C)
	LinkGreen    = color.New(color.FgGreen).Add(color.Bold).SprintFunc()      // #00B174 (PMS Bright Green C)
	ElectricLime = color.New(color.FgHiGreen).SprintFunc()                    // #61FFB6 (PMS 3375)
)

// colorizeValue applies appropriate color to a value based on its type
func colorizeValue(val string, header string, noColor bool) string {
	if noColor {
		return val
	}

	// Status fields (green/yellow/red with increased contrast)
	if header == "status" || header == "provisioning_status" || strings.Contains(header, "state") {
		return colorizeStatus(val, noColor)
	} else if strings.HasSuffix(header, "uid") || strings.HasSuffix(header, "id") {
		// UID fields (use high contrast blue for better visibility of entity identifiers)
		return DodgerBlue(val)
	} else if strings.Contains(header, "price") || strings.Contains(header, "cost") ||
		strings.Contains(header, "rate") {
		// Price/rate fields (use GoldYellow for financial values)
		return GoldYellow(val)
	} else if header == "name" || header == "product_name" || header == "title" {
		// Name fields (use RadRed for emphasis with high contrast for important fields)
		return RadRed(val)
	} else if strings.Contains(header, "speed") || strings.Contains(header, "bandwidth") {
		// Speed/bandwidth fields (use DodgerBlue for technical metrics)
		return DodgerBlue(val)
	} else if header == "location_id" || header == "locationid" || header == "metro" || header == "country" {
		// Location-related fields (use Teal for geographical context)
		return Teal(val)
	} else if strings.Contains(header, "ip") || strings.Contains(header, "cidr") || strings.Contains(header, "subnet") {
		// IP/Network related fields
		return LinkGreen(val)
	} else if strings.Contains(header, "vlan") || strings.Contains(header, "asn") {
		// VLAN/ASN technical fields
		return SkyBlue(val)
	} else if strings.Contains(header, "type") || strings.Contains(header, "product") {
		// Product type fields (use RadRed for key product information)
		return RadRed(val)
	} else if strings.Contains(header, "port") || strings.Contains(header, "interface") {
		// Port/Interface related fields
		return PurpleCloud(val)
	} else if val == "" || val == "<nil>" || val == "null" {
		// Empty values (subtle gray to de-emphasize)
		return color.New(color.FgHiBlack).Sprint("<empty>")
	} else if val == "true" {
		// Boolean true values (LinkGreen for positive)
		return LinkGreen(val)
	} else if val == "false" {
		// Boolean false values (SunsetOrange for negative)
		return SunsetOrange(val)
	}

	// Default value (use a higher contrast color for better readability of regular text)
	return color.New(color.FgWhite).Add(color.Bold).Sprint(val)
}

// Enhance colorizeStatus for better state indication
func colorizeStatus(status string, noColor bool) string {
	if noColor {
		return status
	}

	status = strings.ToUpper(status)

	// Active states - use LinkGreen (success/active)
	if strings.Contains(status, "ACTIVE") || strings.Contains(status, "LIVE") ||
		strings.Contains(status, "CONFIGURED") || status == "UP" || status == "AVAILABLE" {
		return color.New(color.Bold).Sprintf("%s", LinkGreen(status))
	}

	// Warning/transition states - use GoldYellow for warnings/pending
	if strings.Contains(status, "PENDING") || strings.Contains(status, "PROVISIONING") ||
		strings.Contains(status, "WAITING") || strings.Contains(status, "REQUESTED") ||
		strings.Contains(status, "DEPLOYABLE") {
		return color.New(color.Bold).Sprintf("%s", GoldYellow(status))
	}

	// Error/inactive states - use RadRed for errors/critical states (replacing MegaportRed)
	if strings.Contains(status, "ERROR") || strings.Contains(status, "FAILED") {
		return color.New(color.Bold).Sprintf("%s", RadRed(status))
	}

	// Cancelled/deleted states - use SunsetOrange for less severe error states
	if strings.Contains(status, "CANCELLED") || strings.Contains(status, "DELETED") ||
		status == "DOWN" || strings.Contains(status, "INACTIVE") ||
		strings.Contains(status, "DECOMMISSIONING") || strings.Contains(status, "DECOMMISSIONED") {
		return color.New(color.Bold).Sprintf("%s", SunsetOrange(status))
	}

	// Minor warning states - use Mauve for soft warnings
	if strings.Contains(status, "DEGRADED") || strings.Contains(status, "PARTIAL") {
		return color.New(color.Bold).Sprintf("%s", Mauve(status))
	}

	// Reserved states - use Magenta for special states
	if strings.Contains(status, "RESERVED") || strings.Contains(status, "LOCKED") {
		return color.New(color.Bold).Sprintf("%s", Magenta(status))
	}

	// Default for unknown statuses - use a high contrast blue for better visibility
	return color.New(color.Bold).Sprintf("%s", DodgerBlue(status))
}
