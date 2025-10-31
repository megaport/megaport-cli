//go:build js && wasm
// +build js,wasm

package output

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"syscall/js"

	prettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// WasmTableWriter is a global buffer for capturing table output in WASM
var WasmTableWriter = &bytes.Buffer{}

// calculateDynamicWidth calculates the maximum width for a column
func calculateDynamicWidth(termWidth int, minWidth, maxPercentage int) int {
	maxWidth := termWidth * maxPercentage / 100
	if maxWidth < minWidth {
		maxWidth = minWidth
	}
	return maxWidth
}

// printTable is the WASM-specific implementation that properly captures table output
func printTable[T OutputFields](data []T, noColor bool) error {
	headers, fieldIndices, err := getStructTypeInfo(data)
	if err != nil {
		return err
	}
	if len(headers) == 0 {
		return nil
	}

	// Create table writer
	t := prettytable.NewWriter()

	// CRITICAL FIX: In WASM, write ONLY to WasmTableWriter
	// Don't write to os.Stdout as it causes capture issues
	WasmTableWriter.Reset() // Clear previous content
	t.SetOutputMirror(WasmTableWriter)

	js.Global().Get("console").Call("log", "📊 Table will write to WasmTableWriter")

	// WASM-specific table configuration with fixed column widths
	// This ensures consistent, even column distribution in the browser
	columnConfigs := make([]prettytable.ColumnConfig, len(headers))
	for i, header := range headers {
		headerLower := strings.ToLower(header)
		var widthMax int

		// Set specific widths for each column type to match CLI proportions
		switch headerLower {
		case "id":
			widthMax = 6
		case "name", "title":
			widthMax = 35
		case "country":
			widthMax = 16
		case "metro", "city":
			widthMax = 16
		case "site code", "code":
			widthMax = 12
		case "status", "state":
			widthMax = 12
		default:
			widthMax = 20
		}

		columnConfigs[i] = prettytable.ColumnConfig{
			Number:    i + 1,
			WidthMax:  widthMax,
			WidthMin:  widthMax, // Set min = max for consistent width
			AutoMerge: false,
		}
	}
	t.SetColumnConfigs(columnConfigs)

	// WASM with xterm.js: Enable colors since xterm.js supports ANSI codes
	// xterm.js properly renders ANSI color codes and box-drawing characters
	// Keep the noColor parameter from the command flag, don't force it
	// noColor = true  // REMOVED: xterm.js supports colors!

	if noColor {
		// Use a clean, simple style without colors for WASM
		t.SetStyle(prettytable.StyleLight)
	} else {
		// Enhanced Megaport style with prominent headers for WASM
		megaportStyle := prettytable.Style{
			Name: "MegaportEnhancedStyle",
			Box: prettytable.BoxStyle{
				BottomLeft:       "└",
				BottomRight:      "┘",
				BottomSeparator:  "┴",
				Left:             "│",
				LeftSeparator:    "├",
				MiddleHorizontal: "─",
				MiddleSeparator:  "┼",
				MiddleVertical:   "│",
				PaddingLeft:      "  ",  // More padding for better readability
				PaddingRight:     "  ",
				Right:            "│",
				RightSeparator:   "┤",
				TopLeft:          "┌",
				TopRight:         "┐",
				TopSeparator:     "┬",
				UnfinishedRow:    " ≡",
			},
			Color: prettytable.ColorOptions{
				// Bright cyan header with white text for maximum visibility
				Header:       text.Colors{text.FgHiWhite, text.BgCyan, text.Bold},
				Row:          text.Colors{text.FgWhite},  // White text for rows
				RowAlternate: text.Colors{text.FgHiCyan}, // Alternating cyan text
				Footer:       text.Colors{text.FgHiWhite, text.BgBlue, text.Bold},
				Border:       text.Colors{text.FgHiCyan},  // Bright cyan borders
			},
			Format: prettytable.FormatOptions{
				Footer: text.FormatDefault,
				Header: text.FormatUpper,  // Force uppercase headers
				Row:    text.FormatDefault,
			},
			Options: prettytable.Options{
				DrawBorder:      true,
				SeparateColumns: true,
				SeparateFooter:  true,
				SeparateHeader:  true,
				SeparateRows:    false,
			},
		}
		t.SetStyle(megaportStyle)
	}

	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateHeader = true

	headerRow := prettytable.Row{}
	for _, header := range headers {
		headerRow = append(headerRow, strings.ToUpper(header))
	}
	t.AppendHeader(headerRow)

	for _, item := range data {
		if reflect.ValueOf(item).IsZero() {
			continue
		}
		values := extractRowData(item, fieldIndices)
		row := prettytable.Row{}
		for i, val := range values {
			if !noColor {
				val = colorizeValue(val, strings.ToLower(headers[i]), noColor)
			}
			row = append(row, val)
		}
		t.AppendRow(row)
	}

	js.Global().Get("console").Call("log", "🎨 About to render table...")
	t.Render()

	// Get the rendered table output
	tableOutput := WasmTableWriter.String()
	js.Global().Get("console").Call("log", fmt.Sprintf("✅ Table rendered, buffer size: %d bytes", len(tableOutput)))

	// Write the table output to stdout so it can be captured by wasm buffers
	// This is the key: write the buffered content to stdout
	fmt.Print(tableOutput)

	// Also write to a JavaScript-accessible global variable
	js.Global().Set("wasmTableOutput", tableOutput)
	js.Global().Get("console").Call("log", "📝 Table output also stored in wasmTableOutput global")

	return nil
}
