package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	prettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Output interface {
	isOuput()
}

// OutputFields is a marker interface for output-formattable types
type OutputFields interface {
	any
}

// PrintOutput formats data in the specified output style
func PrintOutput[T OutputFields](data []T, format string, noColor bool) error {
	validFormats := map[string]bool{
		"table": true,
		"json":  true,
		"csv":   true,
	}

	if !validFormats[format] {
		return fmt.Errorf("invalid output format: %s", format)
	}

	switch format {
	case "json":
		return printJSON(data)
	case "csv":
		return printCSV(data)
	default:
		if UsePrettyTables {
			return printPrettyTable(data, noColor)
		}
		return printTable(data, noColor) // Original implementation for tests
	}
}

// printJSON outputs formatted JSON to stdout
func printJSON[T OutputFields](data []T) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printTable formats data as a columnar table
func printTable[T OutputFields](data []T, noColor bool) error {
	// Get field information
	headers, fieldIndices, err := getStructTypeInfo(data)
	if err != nil {
		return err
	}

	// Nothing to show
	if len(headers) == 0 {
		return nil
	}

	// Gather all data values in a 2D grid
	rows := collectTableData(data, headers, fieldIndices)

	// Calculate column widths
	colWidths := calculateColumnWidths(rows)

	// Print headers
	printTableHeaders(headers, colWidths, noColor)

	// Print data rows
	printTableRows(rows, headers, colWidths, noColor)

	return nil
}

// getStructTypeInfo extracts type information from the data
func getStructTypeInfo[T OutputFields](data []T) ([]string, []int, error) {
	// Get a sample item to determine fields
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}

	// Get type information via reflection
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		fmt.Println("")
		return nil, nil, nil
	}

	itemType := sampleVal.Type()
	if itemType.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			if itemType.Elem().Kind() != reflect.Struct {
				return nil, nil, nil
			}
			itemType = itemType.Elem()
		} else {
			sampleVal = sampleVal.Elem()
			itemType = sampleVal.Type()
		}
	}

	// Only works with struct types
	if itemType.Kind() != reflect.Struct {
		return nil, nil, nil
	}

	// Extract column headers and field indices
	headers, fieldIndices := extractFieldInfo(itemType)
	return headers, fieldIndices, nil
}

// extractFieldInfo extracts field information from a struct type
func extractFieldInfo(itemType reflect.Type) ([]string, []int) {
	var headers []string
	var fieldIndices []int

	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Validate field type is compatible with output
		if !isOutputCompatibleType(field.Type) {
			continue
		}

		// Look for header tag first
		headerTag := field.Tag.Get("header")
		if headerTag == "-" {
			continue
		}

		// Fall back to csv tag
		if headerTag == "" {
			headerTag = field.Tag.Get("csv")
			if headerTag == "-" {
				continue
			}
		}

		// Fall back to json tag
		if headerTag == "" {
			headerTag = field.Tag.Get("json")
			if headerTag == "-" {
				continue
			}
		}

		// If no tags found, use the field name itself
		if headerTag == "" {
			headerTag = field.Name
		}

		headers = append(headers, headerTag)
		fieldIndices = append(fieldIndices, i)
	}

	return headers, fieldIndices
}

// collectTableData gathers all data values in a 2D grid (rows x columns)
func collectTableData[T OutputFields](data []T, headers []string, fieldIndices []int) [][]string {
	rows := make([][]string, 0, len(data)+1)

	// First row is headers
	rows = append(rows, headers)

	// Gather all data values
	for _, item := range data {
		row := extractRowData(item, fieldIndices)
		if row != nil {
			rows = append(rows, row)
		}
	}

	return rows
}

// extractRowData extracts a single row of data
func extractRowData[T OutputFields](item T, fieldIndices []int) []string {
	itemVal := reflect.ValueOf(item)
	if !itemVal.IsValid() {
		return nil
	}

	if itemVal.Kind() == reflect.Ptr {
		if itemVal.IsNil() {
			return nil
		}
		itemVal = itemVal.Elem()
	}

	// Skip if not a struct
	if itemVal.Kind() != reflect.Struct {
		return nil
	}

	row := make([]string, len(fieldIndices))
	for i, idx := range fieldIndices {
		fieldVal := itemVal.Field(idx)

		if !fieldVal.IsValid() || (fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil()) {
			row[i] = ""
			continue
		}

		if fieldVal.CanInterface() {
			row[i] = formatFieldValue(fieldVal)
		} else {
			row[i] = ""
		}
	}

	return row
}

// calculateColumnWidths determines the maximum width of each column
func calculateColumnWidths(rows [][]string) []int {
	if len(rows) == 0 {
		return nil
	}

	colCount := len(rows[0])
	colWidths := make([]int, colCount)

	for _, row := range rows {
		for i, val := range row {
			if i < colCount && len(val) > colWidths[i] {
				colWidths[i] = len(val)
			}
		}
	}

	return colWidths
}

// printTableHeaders prints the header row with formatting
func printTableHeaders(headers []string, colWidths []int, noColor bool) {
	var headerStrings []string
	for _, header := range headers {
		if !noColor {
			headerStrings = append(headerStrings, color.New(color.Bold).Sprint(header))
		} else {
			headerStrings = append(headerStrings, header)
		}
	}

	fmt.Println(formatRow(headerStrings, colWidths))
}

// printTableRows prints the data rows with colors based on content type
func printTableRows(rows [][]string, headers []string, colWidths []int, noColor bool) {
	// Skip header row (index 0)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		coloredRow := colorizeRow(row, headers, noColor)
		fmt.Println(formatRow(coloredRow, colWidths))
	}
}

// colorizeRow applies appropriate colors to each cell in a row
func colorizeRow(row []string, headers []string, noColor bool) []string {
	coloredRow := make([]string, len(row))

	for j, val := range row {
		coloredVal := val

		// Apply colorization based on field type
		if j < len(headers) {
			header := strings.ToLower(headers[j])
			coloredVal = colorizeValue(val, header, noColor)
		}

		coloredRow[j] = coloredVal
	}

	return coloredRow
}

// colorizeValue applies appropriate color to a value based on its type
func colorizeValue(val string, header string, noColor bool) string {
	if noColor {
		return val
	}

	// Status fields (green/yellow/red with increased contrast)
	if header == "status" || header == "provisioning_status" || strings.Contains(header, "state") {
		return colorizeStatus(val, noColor)
	} else if strings.HasSuffix(header, "uid") || strings.HasSuffix(header, "id") {
		// UID fields (bright cyan for better contrast on dark terminals)
		return color.New(color.FgHiCyan).Sprint(val)
	} else if strings.Contains(header, "price") || strings.Contains(header, "cost") ||
		strings.Contains(header, "rate") {
		// Price/rate fields (magenta - fitting for financial values)
		return color.New(color.FgHiMagenta).Sprint(val)
	} else if header == "name" || header == "product_name" || header == "title" {
		// Name fields (bold white for emphasis with good contrast)
		return color.New(color.FgHiWhite, color.Bold).Sprint(val)
	} else if strings.Contains(header, "speed") || strings.Contains(header, "bandwidth") {
		// Speed/bandwidth fields (bright yellow for attention)
		return color.New(color.FgHiYellow).Sprint(val)
	} else if header == "location_id" || header == "locationid" || header == "metro" || header == "country" {
		// Location-related fields (blue for geographical context)
		return color.New(color.FgBlue).Sprint(val)
	} else if val == "" || val == "<nil>" || val == "null" {
		// Empty values (subtle gray to de-emphasize)
		return color.New(color.FgHiBlack).Sprint("<empty>")
	} else if val == "true" {
		// Boolean true values (green for positive)
		return color.New(color.FgGreen).Sprint(val)
	} else if val == "false" {
		// Boolean false values (red for negative)
		return color.New(color.FgRed).Sprint(val)
	}

	return val
}

// Enhance colorizeStatus for better state indication
func colorizeStatus(status string, noColor bool) string {
	if noColor {
		return status
	}

	status = strings.ToUpper(status)

	// Active states
	if strings.Contains(status, "ACTIVE") || strings.Contains(status, "LIVE") ||
		strings.Contains(status, "CONFIGURED") || status == "UP" || status == "AVAILABLE" {
		return color.New(color.FgGreen, color.Bold).Sprint(status)
	}

	// Warning/transition states
	if strings.Contains(status, "PENDING") || strings.Contains(status, "PROVISIONING") ||
		strings.Contains(status, "WAITING") || strings.Contains(status, "REQUESTED") || strings.Contains(status, "DEPLOYABLE") {
		return color.New(color.FgYellow, color.Bold).Sprint(status)
	}

	// Error/inactive states
	if strings.Contains(status, "ERROR") || strings.Contains(status, "FAILED") ||
		strings.Contains(status, "CANCELLED") || strings.Contains(status, "DELETED") ||
		status == "DOWN" || strings.Contains(status, "INACTIVE") || strings.Contains(status, "DECOMMISSIONING") || strings.Contains(status, "DECOMMISSIONED") {
		return color.New(color.FgRed, color.Bold).Sprint(status)
	}

	// Default for unknown statuses
	return color.New(color.FgHiWhite).Sprint(status)
}

// formatRow formats a row of values with proper spacing based on column widths
func formatRow(values []string, colWidths []int) string {
	var parts []string

	for i, val := range values {
		if i == len(values)-1 {
			// Don't pad the last column
			parts = append(parts, val)
		} else {
			// Calculate visual width (strip ANSI color codes for width calculation)
			// Regular expression to remove ANSI color codes
			re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
			visibleVal := re.ReplaceAllString(val, "")

			// Pad with spaces to match column width + spacing
			padding := colWidths[i] - len(visibleVal) + 3 // Add 3 spaces between columns
			parts = append(parts, val+strings.Repeat(" ", padding))
		}
	}

	return strings.Join(parts, "")
}

// printCSV outputs data in comma-separated value format
// Uses struct tags for column names: csv > json
func printCSV[T OutputFields](data []T) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Get a sample item to determine fields
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}

	// Get type information via reflection
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		// For empty values, just write empty header
		return nil
	}

	t := sampleVal.Type()
	if t.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			// Handle nil pointer
			// Check if we can safely get element type
			if t.Elem().Kind() != reflect.Struct {
				return nil
			}
			t = t.Elem()
			// Create a new instance to inspect fields
			_ = reflect.New(t).Elem()
		} else {
			sampleVal = sampleVal.Elem()
			t = sampleVal.Type()
		}
	}

	// Only works with struct types
	if t.Kind() != reflect.Struct {
		return nil
	}

	// Extract column headers and field names
	var headers []string
	var fields []string
	var fieldIndices []int // Store field indices for safer access

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Validate field type is compatible with output
		if !isOutputCompatibleType(field.Type) {
			continue
		}

		// Look for csv tag first
		csvTag := field.Tag.Get("csv")
		if csvTag == "-" {
			continue // Skip this field
		}

		// Fall back to json tag
		if csvTag == "" {
			csvTag = field.Tag.Get("json")
			if csvTag == "" || csvTag == "-" {
				continue
			}
		}

		headers = append(headers, csvTag)
		fields = append(fields, field.Name)
		fieldIndices = append(fieldIndices, i) // Store actual index
	}

	// Nothing to show
	if len(headers) == 0 {
		return nil
	}

	// Write header row
	if err := w.Write(headers); err != nil {
		return err
	}

	// Write data rows
	for _, item := range data {
		v := reflect.ValueOf(item)
		if !v.IsValid() {
			continue
		}

		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				continue
			}
			v = v.Elem()
		}

		// Skip if not a struct
		if v.Kind() != reflect.Struct {
			continue
		}

		var row []string
		for i, field := range fields {
			// Try direct field access by index first (faster and safer)
			var fieldVal reflect.Value
			if i < len(fieldIndices) && fieldIndices[i] < v.NumField() {
				fieldVal = v.Field(fieldIndices[i])
			} else {
				// Fall back to name-based lookup
				fieldVal = v.FieldByName(field)
			}

			if !fieldVal.IsValid() {
				row = append(row, "")
				continue
			}

			// Handle the case where we can't interface (unexported)
			valueStr := ""
			if fieldVal.CanInterface() {
				// Handle nil interface values
				if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
					row = append(row, "")
					continue
				}

				// Format value based on kind
				valueStr = formatFieldValue(fieldVal)
			}
			row = append(row, valueStr)
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to check if a field type is compatible with output
func isOutputCompatibleType(t reflect.Type) bool {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Basic supported types
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	case reflect.Struct, reflect.Interface:
		// These can be complex but we allow them
		return true
	case reflect.Slice, reflect.Array, reflect.Map:
		// Complex types that may need special handling
		return true
	default:
		// Skip types that don't convert well to string representation
		return false
	}
}

// Helper function to format a field value based on its kind
func formatFieldValue(v reflect.Value) string {
	// Do special handling for time.Time, maps, slices, etc.
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return fmt.Sprintf("%v", v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Struct:
		// Handle common types like time.Time
		if v.Type().String() == "time.Time" {
			if method := v.MethodByName("Format"); method.IsValid() {
				args := []reflect.Value{reflect.ValueOf("2006-01-02")}
				result := method.Call(args)
				if len(result) > 0 {
					return result[0].String()
				}
			}
		}
		return fmt.Sprintf("%v", v.Interface())
	case reflect.Map, reflect.Slice, reflect.Array:
		// For complex types, use json marshaling
		if bytes, err := json.Marshal(v.Interface()); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// captureOutput captures and returns any output written to stdout during execution of f.
func CaptureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

// captureOutputErr is a helper function to capture stdout output and return any error
func CaptureOutputErr(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()
	if err != nil {
		return "", err
	}

	w.Close()
	var buf strings.Builder
	_, err = io.Copy(&buf, r)
	if err != nil {
		return "", err
	}
	os.Stdout = old

	return buf.String(), err
}

// UsePrettyTables controls whether to use the enhanced table rendering
// Default is false for backward compatibility with tests
var UsePrettyTables = false

// printPrettyTable formats data as a columnar table using go-pretty
func printPrettyTable[T OutputFields](data []T, noColor bool) error {
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

	// Choose style based on color preference
	if noColor {
		t.SetStyle(prettytable.StyleLight)
	} else {
		// Create custom Megaport-themed style
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
				UnfinishedRow:    " ≈",
			},
			Color: prettytable.ColorOptions{
				// Megaport red for headers
				Header: text.Colors{text.FgHiRed, text.Bold},
				// White text on normal background for rows
				Row: text.Colors{text.FgHiWhite},
				// Alternate row color for better readability
				RowAlternate: text.Colors{text.FgWhite},
				// Borders in a subtle gray
				Border: text.Colors{text.FgBlack},
				// Footer in the Megaport red for visual balance
				Footer: text.Colors{text.FgHiRed, text.Bold},
				// Special styling for separation row
				Separator: text.Colors{text.FgRed},
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

	// Set common options
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateHeader = true

	// Add header row
	headerRow := prettytable.Row{}
	for _, header := range headers {
		headerRow = append(headerRow, header)
	}
	t.AppendHeader(headerRow)

	// Add data rows using existing row extraction with colorization
	for _, item := range data {
		if reflect.ValueOf(item).IsZero() {
			continue // Skip nil items
		}

		values := extractRowData(item, fieldIndices)
		row := prettytable.Row{}

		// Add values to the row
		for i, val := range values {
			// In go-pretty, we add values directly to the row
			// Colors are applied through the table style rather than per cell
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
