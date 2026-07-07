//go:build js && wasm

package config

// cliHeaders identifies portal/WASM CLI traffic to the Megaport API.
var cliHeaders = map[string]string{"x-app": "cli-wasm"}
