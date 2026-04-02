package utils

import (
	"errors"
	"fmt"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

// WrapAPIError wraps a megaport SDK error with an actionable message based on
// the HTTP status code. Falls back to the original error if the type cannot be
// extracted.
func WrapAPIError(err error, resourceType, resourceUID string) error {
	if err == nil {
		return nil
	}

	var apiErr *megaport.ErrorResponse
	if errors.As(err, &apiErr) {
		switch apiErr.Response.StatusCode {
		case 404:
			listCmd := strings.ToLower(resourceType) + "s"
			return fmt.Errorf("%s %q not found — run 'megaport-cli %s list' to see available resources: %w",
				resourceType, resourceUID, listCmd, err)
		case 401:
			return fmt.Errorf("authentication failed — check your credentials with 'megaport-cli config view' or set MEGAPORT_ACCESS_KEY/MEGAPORT_SECRET_KEY: %w", err)
		case 403:
			return fmt.Errorf("permission denied — your account may not have access to %s %q: %w",
				resourceType, resourceUID, err)
		case 429:
			retryAfter := apiErr.Response.Header.Get("Retry-After")
			if retryAfter != "" {
				return fmt.Errorf("API rate limit exceeded — retry after %s seconds: %w", retryAfter, err)
			}
			return fmt.Errorf("API rate limit exceeded — please wait before retrying: %w", err)
		}
	}
	return err
}
