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

	fields := getOutputFields()
	query := getOutputQuery()

	// Determine what to encode — fields-filtered maps or raw typed data.
	var toEncode interface{}
	if len(fields) > 0 {
		// When --fields is set, validate field names (even on empty data) and build
		// filtered maps so only selected keys appear in the JSON output.
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
		var err error
		toEncode, err = applyJMESPath(query, toEncode)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(toEncode)
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
		} else {
			sampleVal = sampleVal.Elem()
			t = sampleVal.Type()
		}
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	var headers []string
	var jsonNames []string
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
		jsonTag := field.Tag.Get("json")
		if csvTag == "" {
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			csvTag = jsonTag
		}
		// Derive the json name for --fields matching (strip options).
		jn := jsonTag
		if idx := strings.Index(jn, ","); idx != -1 {
			jn = jn[:idx]
		}
		if jn == "" || jn == "-" {
			jn = strings.ToLower(field.Name)
		}
		headers = append(headers, csvTag)
		jsonNames = append(jsonNames, jn)
		fieldIndices = append(fieldIndices, i)
	}
	if csvFields := getOutputFields(); len(csvFields) > 0 {
		var err error
		headers, _, fieldIndices, err = filterByFields(headers, jsonNames, fieldIndices, csvFields)
		if err != nil {
			return err
		}
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
		row := make([]string, 0, len(fieldIndices))
		for _, idx := range fieldIndices {
			if idx >= v.NumField() {
				row = append(row, "")
				continue
			}
			fieldVal := v.Field(idx)
			if !fieldVal.IsValid() || (fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil()) {
				row = append(row, "")
				continue
			}
			row = append(row, fmt.Sprintf("%v", formatFieldValue(fieldVal)))
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

	// Build field names and indices using json tags.
	// displayName holds the header: tag value for --fields alias matching.
	type xmlField struct {
		name        string // json tag name (used as XML element name)
		displayName string // header: tag value (used for --fields header-alias matching)
		index       int
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
		displayName := field.Tag.Get("header")
		if displayName == "" || displayName == "-" {
			displayName = name
		}
		fields = append(fields, xmlField{name: name, displayName: displayName, index: i})
	}
	if xmlFields := getOutputFields(); len(xmlFields) > 0 {
		// Build parallel slices so filterByFields can operate on them.
		// xmlHeaders uses displayName so header-alias matching works (e.g. "Port Speed").
		xmlHeaders := make([]string, len(fields))
		xmlJSONNames := make([]string, len(fields))
		xmlIndices := make([]int, len(fields))
		for i, f := range fields {
			xmlHeaders[i] = f.displayName
			xmlJSONNames[i] = f.name
			xmlIndices[i] = f.index
		}
		_, _, xmlIndices, err := filterByFields(xmlHeaders, xmlJSONNames, xmlIndices, xmlFields)
		if err != nil {
			return err
		}
		// Re-derive names from original fields map by index lookup.
		nameByIndex := make(map[int]string, len(fields))
		for _, f := range fields {
			nameByIndex[f.index] = f.name
		}
		filtered := make([]xmlField, len(xmlIndices))
		for i, idx := range xmlIndices {
			filtered[i] = xmlField{name: nameByIndex[idx], index: idx}
		}
		fields = filtered
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

// osPipe is a variable so tests can replace it to simulate pipe failures.
var osPipe = os.Pipe

func CaptureOutput(f func()) string {
	old := os.Stdout
	r, w, err := osPipe()
	if err != nil {
		f()
		return ""
	}
	os.Stdout = w
	f()
	w.Close()
	out, _ := io.ReadAll(r)
	r.Close()
	os.Stdout = old
	return string(out)
}

func CaptureOutputErr(f func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		runErr := f()
		return "", runErr
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()
	runErr := f()
	w.Close()
	var buf strings.Builder
	_, copyErr := io.Copy(&buf, r)
	if copyErr != nil {
		return "", copyErr
	}
	// Return captured output even when f returned an error so callers can
	// inspect partial output for diagnostics.
	return buf.String(), runErr
}
