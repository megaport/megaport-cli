package cmd

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

type output interface {
	isOuput()
}

// printOutput prints data in either table, JSON, CSV, or XML format
func printOutput[T any](data []T, format string) error {
	switch format {
	case "json":
		return printJSON(data)
	case "csv":
		return printCSV(data)
	case "xml":
		return printXML(data)
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

// printTable handles table output format
func printTable[T any](data []T) error {
	if len(data) == 0 {
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Get struct fields using reflection
	t := reflect.TypeOf(data[0])
	var headers []string
	var fields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			headers = append(headers, tag)
			fields = append(fields, field.Name)
		}
	}

	// Print headers
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print data rows
	for _, item := range data {
		v := reflect.ValueOf(item)
		var row []string
		for _, field := range fields {
			row = append(row, fmt.Sprintf("%v", v.FieldByName(field)))
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}

// printCSV handles CSV output format
func printCSV[T any](data []T) error {
	if len(data) == 0 {
		return nil
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Get headers from struct tags
	t := reflect.TypeOf(data[0])
	var headers []string
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			headers = append(headers, tag)
			fields = append(fields, field.Name)
		}
	}

	if err := w.Write(headers); err != nil {
		return err
	}

	// Write data rows
	for _, item := range data {
		v := reflect.ValueOf(item)
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

// printXML handles XML output format
func printXML[T any](data []T) error {
	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "  ")
	fmt.Println(`<?xml version="1.0" encoding="UTF-8"?>`)
	return encoder.Encode(struct {
		Items []T `xml:"items"`
	}{data})
}
