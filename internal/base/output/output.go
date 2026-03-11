//go:build !wasm
// +build !wasm

package output

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

func printJSON[T OutputFields](data []T) error {
	// Handle nil slices by ensuring we output an empty array instead of null
	if data == nil {
		data = []T{}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

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

func printCSV[T OutputFields](data []T) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		return nil
	}
	t := sampleVal.Type()
	if t.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			if t.Elem().Kind() != reflect.Struct {
				return nil
			}
			t = t.Elem()
			_ = reflect.New(t).Elem()
		} else {
			sampleVal = sampleVal.Elem()
			t = sampleVal.Type()
		}
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	var headers []string
	var fields []string
	var fieldIndices []int
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if !isOutputCompatibleType(field.Type) {
			continue
		}
		csvTag := field.Tag.Get("csv")
		if csvTag == "-" {
			continue
		}
		if csvTag == "" {
			csvTag = field.Tag.Get("json")
			if csvTag == "" || csvTag == "-" {
				continue
			}
		}
		headers = append(headers, csvTag)
		fields = append(fields, field.Name)
		fieldIndices = append(fieldIndices, i)
	}
	if len(headers) == 0 {
		return nil
	}
	if err := w.Write(headers); err != nil {
		return err
	}
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
		if v.Kind() != reflect.Struct {
			continue
		}
		var row []string
		for i, field := range fields {
			var fieldVal reflect.Value
			if i < len(fieldIndices) && fieldIndices[i] < v.NumField() {
				fieldVal = v.Field(fieldIndices[i])
			} else {
				fieldVal = v.FieldByName(field)
			}
			if !fieldVal.IsValid() {
				row = append(row, "")
				continue
			}
			valueStr := ""
			if fieldVal.CanInterface() {
				if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
					row = append(row, "")
					continue
				}
				val := formatFieldValue(fieldVal)
				valueStr = fmt.Sprintf("%v", val)
			}
			row = append(row, valueStr)
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func printXML[T OutputFields](data []T) error {
	if data == nil {
		data = []T{}
	}

	// Determine struct type and field info
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		fmt.Fprint(os.Stdout, xml.Header+"<items></items>\n")
		return nil
	}
	t := sampleVal.Type()
	if t.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			if t.Elem().Kind() != reflect.Struct {
				fmt.Fprint(os.Stdout, xml.Header+"<items></items>\n")
				return nil
			}
			t = t.Elem()
		} else {
			sampleVal = sampleVal.Elem()
			t = sampleVal.Type()
		}
	}
	if t.Kind() != reflect.Struct {
		fmt.Fprint(os.Stdout, xml.Header+"<items></items>\n")
		return nil
	}

	// Build field names and indices using json tags
	type xmlField struct {
		name  string
		index int
	}
	var fields []xmlField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if !isOutputCompatibleType(field.Type) {
			continue
		}
		name := field.Tag.Get("json")
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Tag.Get("csv")
			if name == "-" {
				continue
			}
		}
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		// Strip json tag options (e.g. "name,omitempty" -> "name")
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		fields = append(fields, xmlField{name: name, index: i})
	}

	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "  ")

	fmt.Fprint(os.Stdout, xml.Header)
	start := xml.StartElement{Name: xml.Name{Local: "items"}}
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}

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
		if v.Kind() != reflect.Struct {
			continue
		}

		itemStart := xml.StartElement{Name: xml.Name{Local: "item"}}
		if err := encoder.EncodeToken(itemStart); err != nil {
			return err
		}

		for _, f := range fields {
			fieldVal := v.Field(f.index)
			valueStr := formatFieldValue(fieldVal)

			elemStart := xml.StartElement{Name: xml.Name{Local: f.name}}
			if err := encoder.EncodeToken(elemStart); err != nil {
				return err
			}
			if err := encoder.EncodeToken(xml.CharData(valueStr)); err != nil {
				return err
			}
			if err := encoder.EncodeToken(elemStart.End()); err != nil {
				return err
			}
		}

		if err := encoder.EncodeToken(itemStart.End()); err != nil {
			return err
		}
	}

	if err := encoder.EncodeToken(start.End()); err != nil {
		return err
	}
	if err := encoder.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout)
	return nil
}

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

func CaptureOutputErr(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()
	err := f()
	w.Close()
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	_, copyErr := io.Copy(&buf, r)
	if copyErr != nil {
		return "", copyErr
	}
	return buf.String(), nil
}
