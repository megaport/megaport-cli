package output

import (
	"os"
	"reflect"
	"strings"

	prettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 100
	}
	return width
}

func calculateDynamicWidth(termWidth int, minWidth, maxPercentage int) int {
	maxWidth := termWidth * maxPercentage / 100
	if maxWidth < minWidth {
		maxWidth = minWidth
	}
	return maxWidth
}

func printTable[T OutputFields](data []T, noColor bool) error {
	headers, fieldIndices, err := getStructTypeInfo(data)
	if err != nil {
		return err
	}
	if len(headers) == 0 {
		return nil
	}
	t := prettytable.NewWriter()
	t.SetOutputMirror(os.Stdout)
	termWidth := getTerminalWidth()
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
	t.Render()
	return nil
}
