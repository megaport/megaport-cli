package output

import (
	"fmt"
	"os"
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

// redirectStdout replaces os.Stdout with /dev/null for the duration of the
// benchmark, avoiding the filesystem overhead of CaptureOutput. Returns a
// cleanup function that restores the original stdout.
func redirectStdout(b *testing.B) func() {
	b.Helper()
	old := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	os.Stdout = devNull
	return func() {
		os.Stdout = old
		devNull.Close()
	}
}

func BenchmarkPrintTable(b *testing.B) {
	data := makeBenchData(1000)
	cleanup := redirectStdout(b)
	defer cleanup()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := printTable(data, true); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrintJSON(b *testing.B) {
	data := makeBenchData(1000)
	cleanup := redirectStdout(b)
	defer cleanup()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := printJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrintCSV(b *testing.B) {
	data := makeBenchData(1000)
	cleanup := redirectStdout(b)
	defer cleanup()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := printCSV(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrintXML(b *testing.B) {
	data := makeBenchData(1000)
	cleanup := redirectStdout(b)
	defer cleanup()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := printXML(data); err != nil {
			b.Fatal(err)
		}
	}
}
