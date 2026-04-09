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
		var err error
		_ = CaptureOutput(func() {
			err = printTable(data, true)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrintJSON(b *testing.B) {
	data := makeBenchData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		_ = CaptureOutput(func() {
			err = printJSON(data)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrintCSV(b *testing.B) {
	data := makeBenchData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		_ = CaptureOutput(func() {
			err = printCSV(data)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrintXML(b *testing.B) {
	data := makeBenchData(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		_ = CaptureOutput(func() {
			err = printXML(data)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
