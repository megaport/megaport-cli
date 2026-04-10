//go:build js && wasm
// +build js,wasm

package output

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"syscall/js"

	prettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// isTerminalCached is always false in WASM (no real terminal).
// Uses atomic.Bool for goroutine safety.
var isTerminalCached atomic.Bool

// IsTerminal returns true if stdout is connected to a terminal. Always false in WASM.
func IsTerminal() bool {
	return isTerminalCached.Load()
}

// SetIsTerminal overrides the cached TTY detection result. Intended for tests.
func SetIsTerminal(val bool) {
	isTerminalCached.Store(val)
}

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
	wasmBufMu.Lock()
	defer wasmBufMu.Unlock()
	headers, jsonNames, fieldIndices, err := getStructTypeInfo(data)
	if err != nil {
		return err
	}
	if wasmFields := getOutputFields(); len(wasmFields) > 0 {
		headers, _, fieldIndices, err = filterByFields(headers, jsonNames, fieldIndices, wasmFields)
		if err != nil {
			return err
		}
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

	// WASM-specific table configuration with improved column widths
	// This ensures consistent, readable column distribution in the browser
	columnConfigs := make([]prettytable.ColumnConfig, len(headers))
	for i, header := range headers {
		headerLower := strings.ToLower(header)
		var widthMax int

		// Set specific widths for each column type optimized for web display
		switch headerLower {
		case "uid", "id":
			widthMax = 38 // Full UUID width for better readability
		case "name", "title":
			widthMax = 30 // Reasonable width for port names
		case "locationid", "location_id", "location id":
			widthMax = 12 // Numeric ID
		case "speed", "port_speed", "port speed":
			widthMax = 10 // Numeric speed
		case "status", "provisioning_status", "provisioning status", "state":
			widthMax = 15 // Status strings
		case "country":
			widthMax = 16
		case "metro", "city":
			widthMax = 16
		case "site code", "code":
			widthMax = 12
		case "type":
			widthMax = 12
		default:
			widthMax = 20
		}

		columnConfigs[i] = prettytable.ColumnConfig{
			Number:    i + 1,
			WidthMax:  widthMax,
			WidthMin:  8, // Minimum width for readability
			AutoMerge: false,
		}
	}
	t.SetColumnConfigs(columnConfigs)

	// WASM with xterm.js: Enable colors since xterm.js supports ANSI codes
	// xterm.js properly renders ANSI color codes and box-drawing characters
	// Keep the noColor parameter from the command flag, don't force it
	// noColor = true  // REMOVED: xterm.js supports colors!

	if noColor {
		// Clean, readable style without colors for WASM
		t.SetStyle(prettytable.StyleLight)
	} else {
		// Enhanced style optimized for web terminal with dark background
		megaportStyle := prettytable.Style{
			Name: "MegaportWebStyle",
			Box: prettytable.BoxStyle{
				BottomLeft:       "└",
				BottomRight:      "┘",
				BottomSeparator:  "┴",
				Left:             "│",
				LeftSeparator:    "├",
				MiddleHorizontal: "─",
				MiddleSeparator:  "┼",
				MiddleVertical:   "│",
				PaddingLeft:      " ", // Single space for compact display
				PaddingRight:     " ",
				Right:            "│",
				RightSeparator:   "┤",
				TopLeft:          "┌",
				TopRight:         "┐",
				TopSeparator:     "┬",
				UnfinishedRow:    " ≡",
			},
			Color: prettytable.ColorOptions{
				// Bright, readable colors for dark terminal background
				Header:       text.Colors{text.FgHiCyan, text.Bold}, // Bright cyan text, no background
				Row:          text.Colors{text.FgHiWhite},           // Bright white for rows
				RowAlternate: text.Colors{text.FgCyan},              // Cyan for alternating rows
				Footer:       text.Colors{text.FgHiCyan, text.Bold}, // Bright cyan footer
				Border:       text.Colors{text.FgBlue},              // Blue borders for subtle frame
			},
			Format: prettytable.FormatOptions{
				Footer: text.FormatDefault,
				Header: text.FormatUpper, // Uppercase headers for clarity
				Row:    text.FormatDefault,
			},
			Options: prettytable.Options{
				DrawBorder:      true,
				SeparateColumns: true,
				SeparateFooter:  true,
				SeparateHeader:  true,
				SeparateRows:    false, // No row separation for compact display
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
		v := reflect.ValueOf(item)
		if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
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

	t.Render()

	// Get the rendered table output
	tableOutput := WasmTableWriter.String()

	// Write the table output to stdout so it can be captured by wasm buffers
	// This is the key: write the buffered content to stdout
	fmt.Print(tableOutput)

	// Also write to a JavaScript-accessible global variable
	js.Global().Set("wasmTableOutput", tableOutput)

	return nil
}
