//go:build !js || !wasm

package config

// cliHeaders identifies native CLI traffic to the Megaport API.
var cliHeaders = map[string]string{"x-app": "cli"}
