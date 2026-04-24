//go:build !js && !wasm
// +build !js,!wasm

package config

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// loginFuncMu guards loginFunc, loginFuncWithOutput, and newUnauthenticatedClientFunc.
var loginFuncMu sync.RWMutex

// GetLoginFunc returns the current login function in a thread-safe manner.
func GetLoginFunc() func(context.Context) (*megaport.Client, error) {
	loginFuncMu.RLock()
	defer loginFuncMu.RUnlock()
	return loginFunc
}

// SetLoginFunc replaces the login function in a thread-safe manner.
func SetLoginFunc(fn func(context.Context) (*megaport.Client, error)) {
	loginFuncMu.Lock()
	defer loginFuncMu.Unlock()
	loginFunc = fn
}

// GetLoginFuncWithOutput returns the current output-aware login function in a thread-safe manner.
func GetLoginFuncWithOutput() func(context.Context, string) (*megaport.Client, error) {
	loginFuncMu.RLock()
	defer loginFuncMu.RUnlock()
	return loginFuncWithOutput
}

// SetLoginFuncWithOutput replaces the output-aware login function in a thread-safe manner.
func SetLoginFuncWithOutput(fn func(context.Context, string) (*megaport.Client, error)) {
	loginFuncMu.Lock()
	defer loginFuncMu.Unlock()
	loginFuncWithOutput = fn
}

// GetNewUnauthenticatedClientFunc returns the current unauthenticated client factory in a thread-safe manner.
func GetNewUnauthenticatedClientFunc() func() (*megaport.Client, error) {
	loginFuncMu.RLock()
	defer loginFuncMu.RUnlock()
	return newUnauthenticatedClientFunc
}

// SetNewUnauthenticatedClientFunc replaces the unauthenticated client factory in a thread-safe manner.
func SetNewUnauthenticatedClientFunc(fn func() (*megaport.Client, error)) {
	loginFuncMu.Lock()
	defer loginFuncMu.Unlock()
	newUnauthenticatedClientFunc = fn
}

// resolveEnvironment determines the target API environment using the following
// priority: --env flag > profile config > MEGAPORT_ENVIRONMENT env var > default (production).
// When --profile is set, the named profile's environment is used as a base,
// but --env flag still overrides it.
// If requireProfile is true, errors are returned when --profile is set but the
// profile cannot be loaded; otherwise profile errors are silently ignored.
// The returned value is always a canonical name: "production", "staging", or "development".
func resolveEnvironment(requireProfile bool) (string, error) {
	var env string

	if utils.ProfileOverride != "" {
		manager, err := NewConfigManager()
		if err != nil {
			if requireProfile {
				return "", fmt.Errorf("failed to load config for profile %q: %w", utils.ProfileOverride, err)
			}
		} else {
			profile, err := manager.GetProfile(utils.ProfileOverride)
			if err != nil {
				if requireProfile {
					return "", fmt.Errorf("profile %q not found. Use 'megaport config list-profiles' to see available profiles", utils.ProfileOverride)
				}
			} else if profile.Environment != "" {
				env = profile.Environment
			}
		}
		// --env flag overrides the profile's environment
		if utils.Env != "" {
			env = utils.Env
		}
		// Fall back to env var if still not set
		if env == "" {
			env = os.Getenv("MEGAPORT_ENVIRONMENT")
		}
	} else {
		if utils.Env != "" {
			env = utils.Env
		} else {
			manager, err := NewConfigManager()
			if err == nil {
				profile, _, err := manager.GetCurrentProfile()
				if err == nil && profile.Environment != "" {
					env = profile.Environment
				}
			}
			if env == "" {
				env = os.Getenv("MEGAPORT_ENVIRONMENT")
			}
		}
	}

	return normalizeEnvironment(env), nil
}

func Login(ctx context.Context) (*megaport.Client, error) {
	return GetLoginFunc()(ctx)
}

func LoginWithOutput(ctx context.Context, outputFormat string) (*megaport.Client, error) {
	return GetLoginFuncWithOutput()(ctx, outputFormat)
}

// loginFunc logs into the Megaport API using the current profile or environment variables.
var loginFunc = func(ctx context.Context) (*megaport.Client, error) {
	return GetLoginFuncWithOutput()(ctx, "")
}

// loginFuncWithOutput logs into the Megaport API using the current profile or environment variables.
var loginFuncWithOutput = func(ctx context.Context, outputFormat string) (*megaport.Client, error) {
	var accessKey, secretKey string

	env, err := resolveEnvironment(false)
	if err != nil {
		return nil, err
	}

	// If --profile flag is set, use that specific profile directly
	if utils.ProfileOverride != "" {
		manager, err := NewConfigManager()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		profile, err := manager.GetProfile(utils.ProfileOverride)
		if err != nil {
			return nil, fmt.Errorf("profile %q not found. Use 'megaport config list-profiles' to see available profiles", utils.ProfileOverride)
		}
		accessKey = profile.AccessKey
		secretKey = profile.SecretKey
	} else {
		// Credential selection: if --env flag is used, prefer env vars over profile
		if utils.Env != "" {
			// --env flag was explicitly set, prioritize environment variables
			accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
			secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

			// If env vars are empty, fall back to profile
			if accessKey == "" || secretKey == "" {
				manager, err := NewConfigManager()
				if err == nil {
					profile, _, err := manager.GetCurrentProfile()
					if err == nil {
						if accessKey == "" {
							accessKey = profile.AccessKey
						}
						if secretKey == "" {
							secretKey = profile.SecretKey
						}
					}
				}
			}
		} else {
			// No --env flag, use original priority: profile > env vars
			manager, err := NewConfigManager()
			if err == nil {
				profile, _, err := manager.GetCurrentProfile()
				if err == nil {
					accessKey = profile.AccessKey
					secretKey = profile.SecretKey
				}
			}

			if accessKey == "" {
				accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
			}
			if secretKey == "" {
				secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
			}
		}
	}

	if accessKey == "" {
		return nil, fmt.Errorf("megaport API access key not provided. Configure an active profile or set MEGAPORT_ACCESS_KEY environment variable")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("megaport API secret key not provided. Configure an active profile or set MEGAPORT_SECRET_KEY environment variable")
	}

	envOpt := environmentOption(env)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	opts := appendLogOpts([]megaport.ClientOpt{megaport.WithCredentials(accessKey, secretKey), envOpt, megaport.WithCustomHeaders(cliHeaders)})
	megaportClient, err := megaport.New(httpClient, opts...)
	if err != nil {
		return nil, err
	}

	spinner := output.PrintLoggingInWithOutput(false, outputFormat)
	_, err = megaportClient.Authorize(ctx)

	if err != nil {
		spinner.Stop()
		return nil, err
	} else {
		// Capitalize the first letter of environment for display
		envDisplay := env
		if len(envDisplay) > 0 {
			envDisplay = strings.ToUpper(envDisplay[:1]) + envDisplay[1:]
		}
		spinner.StopWithSuccess(fmt.Sprintf("Successfully logged in to Megaport %s", envDisplay))
	}

	return megaportClient, nil
}

// newUnauthenticatedClientFunc creates a Megaport API client without authentication.
// Used for public API endpoints (e.g., locations) that don't require credentials.
var newUnauthenticatedClientFunc = func() (*megaport.Client, error) {
	env, err := resolveEnvironment(true)
	if err != nil {
		return nil, err
	}

	envOpt := environmentOption(env)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	opts := appendLogOpts([]megaport.ClientOpt{envOpt, megaport.WithCustomHeaders(cliHeaders)})
	return megaport.New(httpClient, opts...)
}

// appendLogOpts appends HTTP debug logging options to the client option slice
// when --log-http is enabled. Logs go to stderr at DEBUG level with sensitive
// fields (access keys, tokens, and all response bodies) redacted.
func appendLogOpts(opts []megaport.ClientOpt) []megaport.ClientOpt {
	result := append([]megaport.ClientOpt(nil), opts...)
	if utils.LogHTTP {
		inner := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
		handler := &redactingHandler{inner: inner}
		result = append(result, megaport.WithLogHandler(handler), megaport.WithLogResponseBody())
	}
	return result
}

// sensitiveKeys lists slog attribute keys whose values should be redacted.
var sensitiveKeys = map[string]bool{
	"access_key":            true,
	"secret_key":            true,
	"response_body_base_64": true,
	"authorization":         true,
	"x-authorization":       true,
	"access_token":          true,
}

// redactingHandler wraps an slog.Handler to replace sensitive attribute values
// with "[REDACTED]" before passing them to the inner handler.
type redactingHandler struct {
	inner slog.Handler
}

func (h *redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *redactingHandler) Handle(ctx context.Context, r slog.Record) error {
	redacted := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		redacted.AddAttrs(redactAttr(a))
		return true
	})
	return h.inner.Handle(ctx, redacted)
}

func (h *redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cleaned := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		cleaned[i] = redactAttr(a)
	}
	return &redactingHandler{inner: h.inner.WithAttrs(cleaned)}
}

func (h *redactingHandler) WithGroup(name string) slog.Handler {
	return &redactingHandler{inner: h.inner.WithGroup(name)}
}

// redactAttr replaces the value of sensitive attributes with "[REDACTED]".
// For group attributes (like the SDK's "api_request" group), it recurses
// into the group's attributes.
func redactAttr(a slog.Attr) slog.Attr {
	if sensitiveKeys[a.Key] {
		return slog.String(a.Key, "[REDACTED]")
	}
	// Recurse into group attributes
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		cleaned := make([]slog.Attr, len(attrs))
		for i, ga := range attrs {
			cleaned[i] = redactAttr(ga)
		}
		return slog.Group(a.Key, attrsToAny(cleaned)...)
	}
	return a
}

func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, a := range attrs {
		result[i] = a
	}
	return result
}

// NewUnauthenticatedClient creates an unauthenticated Megaport API client.
func NewUnauthenticatedClient() (*megaport.Client, error) {
	return GetNewUnauthenticatedClientFunc()()
}
