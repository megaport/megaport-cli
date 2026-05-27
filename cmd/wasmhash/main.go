// Command wasmhash content-hashes a built wasm artifact and rewrites the served
// index.html to point at the hashed name, so the CDN can serve the wasm immutable
// and never invalidate it (ESD-1272). Run after the wasm is copied into the served
// dir and before wasmcompress, which then compresses the hashed file.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const hashLen = 8

// injectedRe matches the script tag wasmhash writes, so a rebuild replaces it
// rather than stacking a second one.
var injectedRe = regexp.MustCompile(`<script>window\.__MEGAPORT_WASM_URL__=[^<]*</script>\n?`)

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil))[:hashLen], nil
}

// insertHash turns .../megaport.wasm into .../megaport.<hash>.wasm.
func insertHash(path, hash string) string {
	ext := filepath.Ext(path)
	stem := strings.TrimSuffix(path, ext)
	return stem + "." + hash + ext
}

func renameWithHash(path string) (newPath, hash string, err error) {
	hash, err = hashFile(path)
	if err != nil {
		return "", "", err
	}
	newPath = insertHash(path, hash)
	if err = os.Rename(path, newPath); err != nil {
		return "", "", err
	}
	return newPath, hash, nil
}

func wasmScript(url string) string {
	// json.Marshal escapes <, >, and & to \uXXXX, so a stray "</script>" in the
	// name can't close the tag early in the HTML parser. Marshaling a string
	// never errors.
	js, _ := json.Marshal(url)
	return "<script>window.__MEGAPORT_WASM_URL__=" + string(js) + "</script>"
}

func injectWasmURL(indexPath, wasmURL string) error {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	html := string(data)
	snippet := wasmScript(wasmURL)

	if injectedRe.MatchString(html) {
		html = injectedRe.ReplaceAllString(html, snippet+"\n")
	} else {
		i := strings.Index(html, "</head>")
		if i < 0 {
			return fmt.Errorf("%s: no </head> to inject the wasm URL into", indexPath)
		}
		html = html[:i] + snippet + "\n" + html[i:]
	}
	return os.WriteFile(indexPath, []byte(html), 0o644)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: wasmhash <wasm-path> <index.html>")
		os.Exit(2)
	}
	wasmPath, indexPath := os.Args[1], os.Args[2]

	newPath, hash, err := renameWithHash(wasmPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wasmhash: %v\n", err)
		os.Exit(1)
	}
	url := "/" + filepath.Base(newPath)
	if err := injectWasmURL(indexPath, url); err != nil {
		fmt.Fprintf(os.Stderr, "wasmhash: %v\n", err)
		os.Exit(1)
	}

	// stdout is just the hashed path so the build can feed it to wasmcompress.
	fmt.Println(newPath)
	fmt.Fprintf(os.Stderr, "wasmhash: %s (hash %s) -> %s\n", wasmPath, hash, url)
}
