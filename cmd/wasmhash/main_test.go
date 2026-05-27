package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wantHash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:hashLen]
}

func TestHashFileDeterministicAndContentSensitive(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.wasm")
	b := filepath.Join(dir, "b.wasm")
	if err := os.WriteFile(a, []byte("the same bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("the same bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	ha, err := hashFile(a)
	if err != nil {
		t.Fatalf("hashFile(a): %v", err)
	}
	hb, err := hashFile(b)
	if err != nil {
		t.Fatalf("hashFile(b): %v", err)
	}
	if ha != hb {
		t.Fatalf("identical content produced different hashes: %s vs %s", ha, hb)
	}
	if len(ha) != hashLen {
		t.Fatalf("hash length = %d, want %d", len(ha), hashLen)
	}
	if ha != wantHash([]byte("the same bytes")) {
		t.Fatalf("hash %s does not match sha256 prefix", ha)
	}

	// A one-byte change must change the hash (acceptance: new build -> new hash).
	if err := os.WriteFile(b, []byte("the same byteS"), 0o644); err != nil {
		t.Fatal(err)
	}
	hb2, err := hashFile(b)
	if err != nil {
		t.Fatal(err)
	}
	if hb2 == ha {
		t.Fatal("changed content produced the same hash")
	}
}

func TestInsertHash(t *testing.T) {
	got := insertHash(filepath.FromSlash("web/vue-demo/megaport.wasm"), "abcd1234")
	want := filepath.FromSlash("web/vue-demo/megaport.abcd1234.wasm")
	if got != want {
		t.Fatalf("insertHash = %q, want %q", got, want)
	}
}

func TestRenameWithHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "megaport.wasm")
	content := []byte("wasm bytes for hashing")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	newPath, hash, err := renameWithHash(path)
	if err != nil {
		t.Fatalf("renameWithHash: %v", err)
	}
	if hash != wantHash(content) {
		t.Fatalf("hash %s does not match content", hash)
	}
	if filepath.Base(newPath) != "megaport."+hash+".wasm" {
		t.Fatalf("newPath base = %q, want megaport.%s.wasm", filepath.Base(newPath), hash)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("original %s should be gone after rename (stat err: %v)", path, err)
	}
	got, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("read hashed file: %v", err)
	}
	if string(got) != string(content) {
		t.Fatal("hashed file content differs from original")
	}
}

const indexTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>demo</title>
    <script type="module" crossorigin src="/assets/main-abc.js"></script>
  </head>
  <body><div id="app"></div></body>
</html>
`

func TestWasmScriptNeutralizesScriptBreakout(t *testing.T) {
	// A "</script>" in the URL must not close the tag early in the HTML parser.
	got := wasmScript("/megaport.</script><script>alert(1)</script>.wasm")
	if n := strings.Count(got, "</script>"); n != 1 {
		t.Fatalf("payload </script> not neutralized, found %d closing tags: %q", n, got)
	}
	if strings.Contains(got, "<script>alert") {
		t.Fatalf("payload <script> survived: %q", got)
	}
}

func TestInjectWasmURLInsertsBeforeHead(t *testing.T) {
	dir := t.TempDir()
	idx := filepath.Join(dir, "index.html")
	if err := os.WriteFile(idx, []byte(indexTemplate), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := injectWasmURL(idx, "/megaport.abcd1234.wasm"); err != nil {
		t.Fatalf("injectWasmURL: %v", err)
	}
	out, err := os.ReadFile(idx)
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)

	want := `<script>window.__MEGAPORT_WASM_URL__="/megaport.abcd1234.wasm"</script>`
	if !strings.Contains(html, want) {
		t.Fatalf("injected script not found.\n%s", html)
	}
	// Must sit inside <head> (before the closing tag) so it runs before the
	// deferred module script that loads the wasm.
	if strings.Index(html, want) > strings.Index(html, "</head>") {
		t.Fatal("injected script is not before </head>")
	}
}

func TestInjectWasmURLIdempotent(t *testing.T) {
	dir := t.TempDir()
	idx := filepath.Join(dir, "index.html")
	if err := os.WriteFile(idx, []byte(indexTemplate), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := injectWasmURL(idx, "/megaport.aaaaaaaa.wasm"); err != nil {
		t.Fatal(err)
	}
	if err := injectWasmURL(idx, "/megaport.bbbbbbbb.wasm"); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(idx)
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)

	if n := strings.Count(html, "__MEGAPORT_WASM_URL__"); n != 1 {
		t.Fatalf("expected exactly one injected script after re-inject, got %d", n)
	}
	if strings.Contains(html, "aaaaaaaa") {
		t.Fatal("stale wasm URL left behind after re-inject")
	}
	if !strings.Contains(html, "/megaport.bbbbbbbb.wasm") {
		t.Fatal("latest wasm URL missing after re-inject")
	}
}

func TestInjectWasmURLNoHead(t *testing.T) {
	dir := t.TempDir()
	idx := filepath.Join(dir, "index.html")
	if err := os.WriteFile(idx, []byte("<html><body>no head</body></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := injectWasmURL(idx, "/megaport.x.wasm"); err == nil {
		t.Fatal("expected error when there is no </head> to inject into")
	}
}
