//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

func main() {
	port := flag.String("port", "8080", "Port to serve on")
	flag.Parse()

	fs := http.FileServer(http.Dir("."))
	http.Handle("/", addCorsHeaders(fs))

	log.Printf("Starting server on http://localhost:%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func addCorsHeaders(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set WASM mime type
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}

		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Serve file
		fs.ServeHTTP(w, r)
	}
}
