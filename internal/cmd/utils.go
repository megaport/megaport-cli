package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// WrapRunE wraps a RunE function to set SilenceUsage to true if an error occurs and formats the error message.
func WrapRunE(runE func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := runE(cmd, args)
		if err != nil {
			// Prevent usage output if an error occurs
			cmd.SilenceUsage = true
			// Silence duplicate error message
			cmd.SilenceErrors = true

			// Return a formatted error message with additional context
			return fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag", cmd.Name(), err, cmd.Name(), args)
		}
		return nil
	}
}

var prompt = func(msg string) (string, error) {
	if !noColor {
		fmt.Print(color.BlueString(msg))
	} else {
		fmt.Print(msg)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// func confirmPrompt(question string) bool {
// 	var response string

// 	if !noColor {
// 		fmt.Print(color.YellowString("%s [y/N]: ", question))
// 	} else {
// 		fmt.Printf("%s [y/N]: ", question)
// 	}

// 	_, err := fmt.Scanln(&response)
// 	if err != nil {
// 		fmt.Println("Error reading input:", err)
// 		return false // Or handle the error as appropriate for your use case
// 	}

// 	response = strings.ToLower(strings.TrimSpace(response))
// 	return response == "y" || response == "yes"
// }

type output interface {
	isOuput()
}

// OutputFields is a marker interface for output-formattable types
type OutputFields interface {
	any
}

// printOutput formats data in the specified output style
func printOutput[T OutputFields](data []T, format string) error {
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
		return printTable(data)
	}
}

// printJSON outputs formatted JSON to stdout
func printJSON[T OutputFields](data []T) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printTable formats data as a columnar table
// Uses struct tags for column headers: header > csv > json
func printTable[T OutputFields](data []T) error {
	// Get a sample item to determine fields
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}

	// Set up tabwriter for readable columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Get type information via reflection
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		// Instead of error, just return empty table for zero values
		fmt.Fprintln(w, "")
		return nil
	}

	itemType := sampleVal.Type()
	if itemType.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			// Handle nil pointer by examining the underlying type
			// Check if we can safely get element type
			if itemType.Elem().Kind() != reflect.Struct {
				return nil
			}
			itemType = itemType.Elem()
			// Create a new instance to inspect fields
			_ = reflect.New(itemType).Elem()
		} else {
			sampleVal = sampleVal.Elem()
			itemType = sampleVal.Type()
		}
	}

	// Only works with struct types
	if itemType.Kind() != reflect.Struct {
		return nil
	}

	// Extract column headers and field names
	var headers []string
	var fields []string
	var fieldIndices []int // Store field indices for safer access

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
			continue // Skip this field
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
		fields = append(fields, field.Name)
		fieldIndices = append(fieldIndices, i) // Store actual index for direct access
	}

	// Nothing to show
	if len(headers) == 0 {
		return nil
	}

	// Print header row
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print data rows
	for _, item := range data {
		itemVal := reflect.ValueOf(item)
		if !itemVal.IsValid() {
			continue
		}

		if itemVal.Kind() == reflect.Ptr {
			if itemVal.IsNil() {
				continue
			}
			itemVal = itemVal.Elem()
		}

		// Skip if not a struct
		if itemVal.Kind() != reflect.Struct {
			continue
		}

		var row []string
		for i, field := range fields {
			// Try direct field access by index first (faster and safer)
			var fieldVal reflect.Value
			if i < len(fieldIndices) && fieldIndices[i] < itemVal.NumField() {
				fieldVal = itemVal.Field(fieldIndices[i])
			} else {
				// Fall back to name-based lookup
				fieldVal = itemVal.FieldByName(field)
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

				// Apply colorization to status fields
				if headers[i] == "Status" || strings.ToLower(headers[i]) == "status" ||
					headers[i] == "provisioning_status" || strings.Contains(strings.ToLower(headers[i]), "state") {
					valueStr = colorizeStatus(valueStr)
				}
			}
			row = append(row, valueStr)
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return nil
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
func captureOutput(f func()) string {
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
func captureOutputErr(f func() error) (string, error) {
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
