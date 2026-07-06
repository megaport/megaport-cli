//go:build js && wasm

package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	megaport "github.com/megaport/megaportgo"
)

// SessionExpiredMarker is a stable substring embedded in the error text
// returned for a rejected WASM external-token, visible in both the plain-text
// "Error: %v" output and the JSON envelope's message field. A host embedding
// the CLI greps command output for this marker to know the injected token was
// rejected and setAuthToken must be called again; see WASM_README.md.
const SessionExpiredMarker = "MEGAPORT_SESSION_EXPIRED"

// wrapSessionExpiredError detects an API auth failure on the WASM external-token
// path and wraps it in a session-expired error carrying SessionExpiredMarker.
// On that path there are no credentials to refresh with, so a 401/403 means the
// host's injected token was rejected and must be re-injected, not just a normal
// auth misconfiguration. Gated on MEGAPORT_ACCESS_TOKEN so setAuthCredentials-based
// (OAuth) logins in the same WASM binary are unaffected; this file only builds
// into the WASM binary in the first place, so native credential logins can
// never reach this code at all.
func wrapSessionExpiredError(err error) error {
	if err == nil || os.Getenv("MEGAPORT_ACCESS_TOKEN") == "" {
		return err
	}
	var apiErr *megaport.ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Response == nil {
		return err
	}
	switch apiErr.Response.StatusCode {
	case 401, 403:
		return exitcodes.NewSessionExpiredError(fmt.Errorf("%s: session expired, please re-authenticate: %w", SessionExpiredMarker, err))
	default:
		return err
	}
}
