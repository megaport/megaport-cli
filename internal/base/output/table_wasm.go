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
	
	// In WASM, use a fixed width since term.GetSize() doesn't work
	termWidth := 120
	
	columnConfigs := make([]prettytable.ColumnConfig, len(headers))
	for i := range headers {
		if i == 0 {
			columnConfigs[i] = prettytable.ColumnConfig{Number: i + 1, WidthMax: calculateDynamicWidth(termWidth, 10, 15)}
		} else {
			columnConfigs[i] = prettytable.ColumnConfig{Number: i + 1, WidthMax: calculateDynamicWidth(termWidth, 15, 25)}
		}
	}
	t.SetColumnConfigs(columnConfigs)
	
	if noColor {
		t.SetStyle(prettytable.StyleLight)
	} else {
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
				Header:       text.Colors{text.FgHiWhite, text.BgRed, text.Bold},
				Row:          text.Colors{},
				RowAlternate: text.Colors{text.FgHiBlack},
				Footer:       text.Colors{text.FgHiWhite, text.BgRed, text.Bold},
				Border:       text.Colors{text.FgBlue},
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
