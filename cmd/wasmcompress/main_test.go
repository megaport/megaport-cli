package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andybalholm/brotli"
)

// sampleWASM returns compressible bytes standing in for the real artifact.
func sampleWASM() []byte {
	return bytes.Repeat([]byte("megaport-cli wasm payload 0123456789\n"), 8192)
}

func TestWriteBrotliRoundTrip(t *testing.T) {
	in := sampleWASM()
	var buf bytes.Buffer
	if err := writeBrotli(&buf, bytes.NewReader(in)); err != nil {
		t.Fatalf("writeBrotli: %v", err)
	}
	if buf.Len() >= len(in) {
		t.Fatalf("brotli output not smaller: got %d, input %d", buf.Len(), len(in))
	}
	got, err := io.ReadAll(brotli.NewReader(bytes.NewReader(buf.Bytes())))
	if err != nil {
		t.Fatalf("brotli decode: %v", err)
	}
	if !bytes.Equal(got, in) {
		t.Fatal("brotli round-trip mismatch")
	}
}

func TestWriteGzipRoundTrip(t *testing.T) {
	in := sampleWASM()
	var buf bytes.Buffer
	if err := writeGzip(&buf, bytes.NewReader(in)); err != nil {
		t.Fatalf("writeGzip: %v", err)
	}
	if buf.Len() >= len(in) {
		t.Fatalf("gzip output not smaller: got %d, input %d", buf.Len(), len(in))
	}
	r, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("gzip decode: %v", err)
	}
	if !bytes.Equal(got, in) {
		t.Fatal("gzip round-trip mismatch")
	}
}

func TestCompressWASMProducesObjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "megaport.wasm")
	in := sampleWASM()
	if err := os.WriteFile(path, in, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := compressWASM(path)
	if err != nil {
		t.Fatalf("compressWASM: %v", err)
	}
	if res.rawSize != int64(len(in)) {
		t.Fatalf("rawSize = %d, want %d", res.rawSize, len(in))
	}

	for ext, size := range map[string]int64{".br": res.brSize, ".gz": res.gzSize} {
		fi, err := os.Stat(path + ext)
		if err != nil {
			t.Fatalf("expected %s on disk: %v", ext, err)
		}
		if fi.Size() != size {
			t.Fatalf("%s on-disk size %d != reported %d", ext, fi.Size(), size)
		}
		if size >= res.rawSize {
			t.Fatalf("%s (%d) not smaller than raw (%d)", ext, size, res.rawSize)
		}
	}

	br, err := os.Open(path + ".br")
	if err != nil {
		t.Fatal(err)
	}
	defer br.Close()
	got, err := io.ReadAll(brotli.NewReader(br))
	if err != nil {
		t.Fatalf("brotli decode: %v", err)
	}
	if !bytes.Equal(got, in) {
		t.Fatal("brotli object did not round-trip to original wasm")
	}
}

func TestCompressWASMMissingFile(t *testing.T) {
	if _, err := compressWASM(filepath.Join(t.TempDir(), "nope.wasm")); err == nil {
		t.Fatal("expected error for missing input file")
	}
}
