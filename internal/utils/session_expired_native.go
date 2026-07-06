//go:build !js || !wasm

package utils

// SessionExpiredMarker is a stable substring embedded in the error text
// returned for a rejected WASM external-token, visible in both the plain-text
// "Error: %v" output and the JSON envelope's message field. A host embedding
// the CLI greps command output for this marker to know the injected token was
// rejected and setAuthToken must be called again; see WASM_README.md. Defined
// here too (native builds never emit it) purely so tests and other packages
// have one symbol to reference regardless of build target.
const SessionExpiredMarker = "MEGAPORT_SESSION_EXPIRED"

// wrapSessionExpiredError is a no-op on native builds. The WASM
// external-token session-expiry signal only makes sense when the CLI is
// running as WASM in a browser with no credentials to fall back on; native
// credential-auth failures always pass through unchanged.
func wrapSessionExpiredError(err error) error {
	return err
}
