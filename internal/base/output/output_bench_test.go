package output

import (
	"fmt"
	"testing"
)

type benchStruct struct {
	ID     int    `json:"id" csv:"id" header:"ID"`
	Name   string `json:"name" csv:"name" header:"Name"`
	Status string `json:"status" csv:"status" header:"Status"`
	Active bool   `json:"active" csv:"active" header:"Active"`
}

func makeBenchData(n int) []benchStruct { //nolint:unparam // n is parameterized for reuse
	data := make([]benchStruct, n)
	for i := range data {
		data[i] = benchStruct{
			ID:     i,
			Name:   fmt.Sprintf("item-%d", i),
			Status: "active",
			Active: i%2 == 0,
		}
	}
	return data
}

func BenchmarkPrintTable(b *testing.B) {
	data := makeBenchData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CaptureOutput(func() {
			_ = printTable(data, true)
		})
	}
}

func BenchmarkPrintJSON(b *testing.B) {
	data := makeBenchData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CaptureOutput(func() {
			_ = printJSON(data)
		})
	}
}

func BenchmarkPrintCSV(b *testing.B) {
	data := makeBenchData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CaptureOutput(func() {
			_ = printCSV(data)
		})
	}
}

func BenchmarkPrintXML(b *testing.B) {
	data := makeBenchData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CaptureOutput(func() {
			_ = printXML(data)
		})
	}
}
