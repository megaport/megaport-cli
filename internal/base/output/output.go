package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type Output interface {
	isOuput()
}

type OutputFields interface {
	any
}

type ResourceTag struct {
	Key   string `json:"key" header:"KEY"`
	Value string `json:"value" header:"VALUE"`
}

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
		return printTable(data, noColor)
	}
}

func printJSON[T OutputFields](data []T) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func getStructTypeInfo[T OutputFields](data []T) ([]string, []int, error) {
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}
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
	if itemType.Kind() != reflect.Struct {
		return nil, nil, nil
	}
	headers, fieldIndices := extractFieldInfo(itemType)
	return headers, fieldIndices, nil
}

func extractFieldInfo(itemType reflect.Type) ([]string, []int) {
	var headers []string
	var fieldIndices []int
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if !isOutputCompatibleType(field.Type) {
			continue
		}
		headerTag := field.Tag.Get("header")
		if headerTag == "-" {
			continue
		}
		if headerTag == "" {
			headerTag = field.Tag.Get("csv")
			if headerTag == "-" {
				continue
			}
		}
		if headerTag == "" {
			headerTag = field.Tag.Get("json")
			if headerTag == "-" {
				continue
			}
		}
		if headerTag == "" {
			headerTag = field.Name
		}
		headers = append(headers, headerTag)
		fieldIndices = append(fieldIndices, i)
	}
	return headers, fieldIndices
}

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

func isOutputCompatibleType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	case reflect.Struct, reflect.Interface:
		return true
	case reflect.Slice, reflect.Array, reflect.Map:
		return true
	default:
		return false
	}
}

func formatFieldValue(v reflect.Value) string {
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
		if bytes, err := json.Marshal(v.Interface()); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
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
