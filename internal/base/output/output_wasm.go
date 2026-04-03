//go:build js && wasm
// +build js,wasm

package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"syscall/js"

	"github.com/jmespath/go-jmespath"
)

// WasmJSONWriter is a global buffer for capturing JSON output in WASM
var WasmJSONWriter = &bytes.Buffer{}

// WasmCSVWriter is a global buffer for capturing CSV output in WASM
var WasmCSVWriter = &bytes.Buffer{}

// WasmXMLWriter is a global buffer for capturing XML output in WASM
var WasmXMLWriter = &bytes.Buffer{}

// printJSON is the WASM-specific implementation that properly captures JSON output
func printJSON[T OutputFields](data []T) error {
	WasmJSONWriter.Reset()

	if data == nil {
		data = []T{}
	}

	fields := getOutputFields()
	query := getOutputQuery()

	// Determine what to encode — fields-filtered maps or raw typed data.
	var toEncode interface{}
	if len(fields) > 0 {
		headers, jsonNames, indices, err := getStructTypeInfo(data)
		if err != nil {
			return err
		}
		_, jsonNames, indices, err = filterByFields(headers, jsonNames, indices, fields)
		if err != nil {
			return err
		}
		rows := make([]interface{}, 0, len(data))
		for _, item := range data {
			v := reflect.ValueOf(item)
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					rows = append(rows, nil)
					continue
				}
				v = v.Elem()
			}
			if !v.IsValid() || v.Kind() != reflect.Struct {
				rows = append(rows, nil)
				continue
			}
			m := make(map[string]interface{}, len(indices))
			for i, idx := range indices {
				if idx >= v.NumField() {
					continue
				}
				m[jsonNames[i]] = v.Field(idx).Interface()
			}
			rows = append(rows, m)
		}
		toEncode = rows
	} else {
		toEncode = data
	}

	// Apply JMESPath query if set.
	if query != "" {
		raw, err := json.Marshal(toEncode)
		if err != nil {
			return err
		}
		var parsed interface{}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return err
		}
		result, err := jmespath.Search(query, parsed)
		if err != nil {
			return fmt.Errorf("invalid JMESPath query %q: %w", query, err)
		}
		toEncode = result
	}

	encoder := json.NewEncoder(WasmJSONWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(toEncode); err != nil {
		js.Global().Get("console").Call("error", "Failed to encode JSON:", err.Error())
		return err
	}

	jsonOutput := WasmJSONWriter.String()
	fmt.Print(jsonOutput)
	js.Global().Set("wasmJSONOutput", jsonOutput)
	return nil
}

// printCSV is the WASM-specific implementation that properly captures CSV output
func printCSV[T OutputFields](data []T) error {
	// Reset the buffer to ensure clean output
	WasmCSVWriter.Reset()

	js.Global().Get("console").Call("log", "📊 CSV will write to WasmCSVWriter")

	// Create CSV writer that writes to our buffer
	w := csv.NewWriter(WasmCSVWriter)
	defer w.Flush()

	// Get the first sample to determine fields
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

	// Build headers and field indices using the same logic as the regular version
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

	// Write header row
	if err := w.Write(headers); err != nil {
		js.Global().Get("console").Call("error", "Failed to write CSV header:", err.Error())
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
			js.Global().Get("console").Call("error", "Failed to write CSV row:", err.Error())
			return err
		}
	}

	// Flush the CSV writer to ensure all data is written to the buffer
	w.Flush()

	// Get the CSV output
	csvOutput := WasmCSVWriter.String()
	js.Global().Get("console").Call("log", fmt.Sprintf("✅ CSV encoded, buffer size: %d bytes", len(csvOutput)))

	// Write the CSV output to stdout so it can be captured by wasm buffers
	fmt.Print(csvOutput)

	// Also write to a JavaScript-accessible global variable for direct access
	js.Global().Set("wasmCSVOutput", csvOutput)
	js.Global().Get("console").Call("log", "📝 CSV output also stored in wasmCSVOutput global")

	return nil
}

// printXML is the WASM-specific implementation that properly captures XML output
func printXML[T OutputFields](data []T) error {
	WasmXMLWriter.Reset()

	js.Global().Get("console").Call("log", "📄 XML will write to WasmXMLWriter")

	if data == nil {
		data = []T{}
	}

	var sample T
	if len(data) > 0 {
		sample = data[0]
	}
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		WasmXMLWriter.WriteString(xml.Header + "<items></items>\n")
		xmlOutput := WasmXMLWriter.String()
		fmt.Print(xmlOutput)
		js.Global().Set("wasmXMLOutput", xmlOutput)
		return nil
	}
	t := sampleVal.Type()
	if t.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			if t.Elem().Kind() != reflect.Struct {
				WasmXMLWriter.WriteString(xml.Header + "<items></items>\n")
				xmlOutput := WasmXMLWriter.String()
				fmt.Print(xmlOutput)
				js.Global().Set("wasmXMLOutput", xmlOutput)
				return nil
			}
			t = t.Elem()
		} else {
			sampleVal = sampleVal.Elem()
			t = sampleVal.Type()
		}
	}
	if t.Kind() != reflect.Struct {
		WasmXMLWriter.WriteString(xml.Header + "<items></items>\n")
		xmlOutput := WasmXMLWriter.String()
		fmt.Print(xmlOutput)
		js.Global().Set("wasmXMLOutput", xmlOutput)
		return nil
	}

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
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		fields = append(fields, xmlField{name: name, index: i})
	}

	encoder := xml.NewEncoder(WasmXMLWriter)
	encoder.Indent("", "  ")

	WasmXMLWriter.WriteString(xml.Header)
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
	WasmXMLWriter.WriteString("\n")

	xmlOutput := WasmXMLWriter.String()
	js.Global().Get("console").Call("log", fmt.Sprintf("✅ XML encoded, buffer size: %d bytes", len(xmlOutput)))
	fmt.Print(xmlOutput)
	js.Global().Set("wasmXMLOutput", xmlOutput)
	js.Global().Get("console").Call("log", "📝 XML output also stored in wasmXMLOutput global")

	return nil
}

// calculateColumnWidths calculates the maximum width for each column
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

// CaptureOutput runs a function and captures its stdout output
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

// CaptureOutputErr runs a function and captures its stdout output, also returning any error
func CaptureOutputErr(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := f()
	if err != nil {
		os.Stdout = old
		return "", err
	}
	w.Close()
	var buf strings.Builder
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String(), nil
}
