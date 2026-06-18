// Command wasmcompress writes brotli (q11) and gzip (-9) copies of a wasm
// artifact next to it. CloudFront only auto-compresses objects up to 10 MB, and
// the wasm is well above that even after stripping glamour, so it must be
// pre-compressed at the origin (ESD-1268).
package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andybalholm/brotli"
)

const defaultWASM = "web/megaport.wasm"

func writeBrotli(dst io.Writer, src io.Reader) error {
	w := brotli.NewWriterLevel(dst, brotli.BestCompression)
	if _, err := io.Copy(w, src); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func writeGzip(dst io.Writer, src io.Reader) error {
	w, err := gzip.NewWriterLevel(dst, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, src); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

type result struct {
	rawSize, brSize, gzSize int64
}

func compressWASM(path string) (result, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return result{}, err
	}
	res := result{rawSize: fi.Size()}
	if res.brSize, err = encodeToFile(path, path+".br", writeBrotli); err != nil {
		return res, fmt.Errorf("brotli: %w", err)
	}
	if res.gzSize, err = encodeToFile(path, path+".gz", writeGzip); err != nil {
		return res, fmt.Errorf("gzip: %w", err)
	}
	return res, nil
}

func encodeToFile(srcPath, dstPath string, encode func(io.Writer, io.Reader) error) (size int64, err error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	// Encode to a temp file and rename on success, so a failed or partial encode
	// never leaves a truncated artifact at dstPath (these get uploaded to a CDN).
	tmp, err := os.CreateTemp(filepath.Dir(dstPath), filepath.Base(dstPath)+".*.tmp")
	if err != nil {
		return 0, err
	}
	tmpName := tmp.Name()
	// On any error, close the handle (so the file can be removed on Windows) and
	// delete the temp file. On the error paths after the explicit Close below this
	// closes a second time and gets ErrClosed, which we ignore.
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	// os.CreateTemp makes the file 0600; served artifacts need to be readable.
	if err = tmp.Chmod(0o644); err != nil {
		return 0, err
	}
	if err = encode(tmp, src); err != nil {
		return 0, err
	}
	if err = tmp.Close(); err != nil {
		return 0, err
	}
	fi, err := os.Stat(tmpName)
	if err != nil {
		return 0, err
	}
	if err = os.Rename(tmpName, dstPath); err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func main() {
	path := defaultWASM
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	res, err := compressWASM(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wasmcompress: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wasmcompress %s\n", path)
	fmt.Printf("  raw     %6.2f MiB\n", mib(res.rawSize))
	fmt.Printf("  brotli  %6.2f MiB  (%s.br, %.0f%% of raw)\n", mib(res.brSize), path, pct(res.brSize, res.rawSize))
	fmt.Printf("  gzip    %6.2f MiB  (%s.gz, %.0f%% of raw)\n", mib(res.gzSize), path, pct(res.gzSize, res.rawSize))
}

func mib(n int64) float64 { return float64(n) / (1024 * 1024) }

func pct(part, whole int64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole) * 100
}
