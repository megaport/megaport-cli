package output

import (
	"os"
	"reflect"
	"strings"

	prettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

// getTerminalWidth returns the current terminal width or a default value if it cannot be determined
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 100 // Default width if terminal size cannot be determined
	}
	return width
}

// calculateDynamicWidth determines column width based on terminal size and content requirements
func calculateDynamicWidth(termWidth int, minWidth, maxPercentage int) int {
	// Calculate maximum width as percentage of terminal width
	maxWidth := termWidth * maxPercentage / 100

	// Ensure width is at least the minimum required
	if maxWidth < minWidth {
		maxWidth = minWidth
	}

	return maxWidth
}

func printTable[T OutputFields](data []T, noColor bool) error {
	// Get field information using existing extraction logic
	headers, fieldIndices, err := getStructTypeInfo(data)
	if err != nil {
		return err
	}

	// Nothing to show if no headers were found
	if len(headers) == 0 {
		return nil
	}

	// Create and configure table
	t := prettytable.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Get terminal width to adjust column sizes accordingly
	termWidth := getTerminalWidth()

	// Set column configurations - more dynamic based on terminal width
	columnConfigs := make([]prettytable.ColumnConfig, len(headers))
	for i := range headers {
		// For first column (typically ID/UID), keep it reasonably compact
		if i == 0 {
			columnConfigs[i] = prettytable.ColumnConfig{Number: i + 1, WidthMax: calculateDynamicWidth(termWidth, 10, 15)}
		} else {
			// For other columns, allow more space but still constrain
			columnConfigs[i] = prettytable.ColumnConfig{Number: i + 1, WidthMax: calculateDynamicWidth(termWidth, 15, 25)}
		}
	}
	t.SetColumnConfigs(columnConfigs)

	// Choose style based on color preference
	if noColor {
		t.SetStyle(prettytable.StyleLight)
	} else {
		// Create custom Megaport-themed style using brand colors
		megaportStyle := prettytable.Style{
			Name: "MegaportStyle",
			Box: prettytable.BoxStyle{
				BottomLeft:       "└",
				BottomRight:      "┘",
				BottomSeparator:  "┴",
				Left:             "│",
				LeftSeparator:    "├",
				MiddleHorizontal: "─",
				MiddleSeparator:  "┼",
				MiddleVertical:   "│",
				PaddingLeft:      " ",
				PaddingRight:     " ",
				Right:            "│",
				RightSeparator:   "┤",
				TopLeft:          "┌",
				TopRight:         "┐",
				TopSeparator:     "┬",
				UnfinishedRow:    " ≡",
			},
			Color: prettytable.ColorOptions{
				// RadRed/DeepNightBlue for headers - core Megaport brand
				Header: text.Colors{text.FgHiWhite, text.BgRed, text.Bold},
				// Normal rows use default color scheme with enhanced readability
				Row: text.Colors{},
				// Alternate rows get subtle distinction for easy reading
				RowAlternate: text.Colors{text.FgHiBlack},
				// Footer matches header styling
				Footer: text.Colors{text.FgHiWhite, text.BgRed, text.Bold},
				// Border in DarkBlue for contrast and brand representation
				Border: text.Colors{text.FgBlue},
			},
			Format: prettytable.FormatOptions{
				Footer: text.FormatDefault,
				Header: text.FormatTitle,
				Row:    text.FormatDefault,
			},
			Options: prettytable.Options{
				DrawBorder:      true,
				SeparateColumns: true,
				SeparateFooter:  true,
				SeparateHeader:  true,
				SeparateRows:    false, // Disable row separators for cleaner look
			},
		}
		t.SetStyle(megaportStyle)
	}

	// Set common options
	t.Style().Options.DrawBorder = true // Enable full border for better readability
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateHeader = true

	// Add header row
	headerRow := prettytable.Row{}
	for _, header := range headers {
		// Convert header to uppercase here
		headerRow = append(headerRow, strings.ToUpper(header))
	}
	t.AppendHeader(headerRow)

	// Add data rows using existing row extraction with colorization
	for _, item := range data {
		if reflect.ValueOf(item).IsZero() {
			continue // Skip nil items
		}

		values := extractRowData(item, fieldIndices)
		row := prettytable.Row{}

		// Add values to the row with colorization
		for i, val := range values {
			// Apply colorization based on column type and value
			if !noColor {
				val = colorizeValue(val, strings.ToLower(headers[i]), noColor)
			}
			row = append(row, val)
		}
		t.AppendRow(row)
	}

	// Render the table
	t.Render()
	return nil
}
