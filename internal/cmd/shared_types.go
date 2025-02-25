package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

type output interface {
	isOuput()
}

func printOutput[T any](data []T, format string) error {
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

// printJSON handles JSON output format
func printJSON[T any](data []T) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printTable prints a header row (from struct tags) even if data is empty.
func printTable[T any](data []T) error {
	// Use the first item if available. Otherwise, create a zero value.
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}

	// Configure tabwriter for left alignment and consistent spacing
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Reflect on sample. If sample is a pointer, dereference it.
	sampleVal := reflect.ValueOf(sample)
	itemType := sampleVal.Type()
	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}

	// Collect struct fields and headers
	var headers []string
	var fields []string

	// If T is not a struct (e.g., empty interface), just return
	if itemType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			headers = append(headers, tag)
			fields = append(fields, field.Name)
		}
	}

	// Print headers with tabs
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print data rows (if any)
	for _, item := range data {
		itemVal := reflect.ValueOf(item)
		if itemVal.Kind() == reflect.Ptr {
			itemVal = itemVal.Elem()
		}

		var row []string
		for _, field := range fields {
			row = append(row, fmt.Sprintf("%v", itemVal.FieldByName(field)))
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}

// printCSV prints a header row (from struct tags) even if data is empty.
func printCSV[T any](data []T) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Use the first item if available. Otherwise, create a zero value.
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}

	sampleVal := reflect.ValueOf(sample)
	t := sampleVal.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// If T is not a struct (e.g., empty interface), just return
	if t.Kind() != reflect.Struct {
		return nil
	}

	// Get headers from struct tags
	var headers []string
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			headers = append(headers, tag)
			fields = append(fields, field.Name)
		}
	}

	// Always print headers
	if err := w.Write(headers); err != nil {
		return err
	}

	// Write data rows if present
	for _, item := range data {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		var row []string
		for _, field := range fields {
			row = append(row, fmt.Sprintf("%v", v.FieldByName(field)))
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}
