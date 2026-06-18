package main

import (
	"bytes"
	"compress/gzip"
	"errors"
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
	if err := r.Close(); err != nil {
		t.Fatalf("gzip close (CRC/footer): %v", err)
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

func TestCompressWASMArtifactsAreReadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "megaport.wasm")
	if err := os.WriteFile(path, sampleWASM(), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := compressWASM(path); err != nil {
		t.Fatalf("compressWASM: %v", err)
	}
	// Served/packaged artifacts must be group- and world-readable, not the
	// 0600 that os.CreateTemp would leave behind.
	for _, ext := range []string{".br", ".gz"} {
		fi, err := os.Stat(path + ext)
		if err != nil {
			t.Fatal(err)
		}
		if perm := fi.Mode().Perm(); perm&0o044 != 0o044 {
			t.Fatalf("%s perm = %04o, want group+other readable", ext, perm)
		}
	}
}

func TestEncodeToFileNoPartialOnError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.wasm")
	if err := os.WriteFile(src, sampleWASM(), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "out.br")

	boom := errors.New("encode blew up")
	_, err := encodeToFile(src, dst, func(w io.Writer, _ io.Reader) error {
		_, _ = w.Write([]byte("half-written")) // partial output before failing
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("got err %v, want %v", err, boom)
	}
	if _, statErr := os.Stat(dst); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected no artifact at %s on encode failure (stat err: %v)", dst, statErr)
	}
	// The temp file must be cleaned up too — only the source should remain.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "src.wasm" {
			t.Fatalf("leftover file after failed encode: %s", e.Name())
		}
	}
}

func TestCompressWASMMissingFile(t *testing.T) {
	if _, err := compressWASM(filepath.Join(t.TempDir(), "nope.wasm")); err == nil {
		t.Fatal("expected error for missing input file")
	}
}
