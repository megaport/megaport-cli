package output

import "time"

// Shared test fixture types used by both the native (!wasm) and WASM test
// builds. They live in an untagged file so neither build redefines them.

type SimpleStruct struct {
	ID     int    `json:"id" csv:"id" header:"ID"`
	Name   string `json:"name" csv:"name" header:"Name"`
	Active bool   `json:"active" csv:"active" header:"Active"`
}

type ComplexStruct struct {
	ID        int               `json:"id" csv:"id" header:"ID"`
	Name      string            `json:"name" csv:"name" header:"Name"`
	Created   time.Time         `json:"created" csv:"created" header:"Created"`
	Tags      []string          `json:"tags" csv:"tags" header:"Tags"`
	Metadata  map[string]string `json:"metadata" csv:"metadata" header:"Metadata"`
	Reference *SimpleStruct     `json:"reference" csv:"reference" header:"Reference"`
	Ignored   int               `json:"-" csv:"-" header:"-"`
}
