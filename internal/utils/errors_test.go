package utils

import (
	"errors"
	"net/http"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func makeAPIError(statusCode int, retryAfter string) *megaport.ErrorResponse {
	header := http.Header{}
	if retryAfter != "" {
		header.Set("Retry-After", retryAfter)
	}
	return &megaport.ErrorResponse{
		Response: &http.Response{
			StatusCode: statusCode,
			Header:     header,
			Request:    &http.Request{},
		},
		Message: "test error",
	}
}

func TestWrapAPIError_Nil(t *testing.T) {
	assert.Nil(t, WrapAPIError(nil, "Port", "uid-123"))
}

func TestWrapAPIError_404(t *testing.T) {
	apiErr := makeAPIError(404, "")
	wrapped := WrapAPIError(apiErr, "Port", "uid-123")
	assert.Contains(t, wrapped.Error(), "not found")
	assert.Contains(t, wrapped.Error(), "ports list")
	assert.Contains(t, wrapped.Error(), "uid-123")
	// Original error still unwrappable
	var target *megaport.ErrorResponse
	assert.True(t, errors.As(wrapped, &target))
}

func TestWrapAPIError_401(t *testing.T) {
	apiErr := makeAPIError(401, "")
	wrapped := WrapAPIError(apiErr, "Port", "uid-123")
	assert.Contains(t, wrapped.Error(), "authentication failed")
	assert.Contains(t, wrapped.Error(), "config view")
}

func TestWrapAPIError_403(t *testing.T) {
	apiErr := makeAPIError(403, "")
	wrapped := WrapAPIError(apiErr, "MCR", "mcr-456")
	assert.Contains(t, wrapped.Error(), "permission denied")
	assert.Contains(t, wrapped.Error(), "mcr-456")
}

func TestWrapAPIError_429_NoRetryAfter(t *testing.T) {
	apiErr := makeAPIError(429, "")
	wrapped := WrapAPIError(apiErr, "Port", "uid-123")
	assert.Contains(t, wrapped.Error(), "rate limit exceeded")
	assert.Contains(t, wrapped.Error(), "please wait")
}

func TestWrapAPIError_429_WithRetryAfter(t *testing.T) {
	apiErr := makeAPIError(429, "30")
	wrapped := WrapAPIError(apiErr, "Port", "uid-123")
	assert.Contains(t, wrapped.Error(), "rate limit exceeded")
	assert.Contains(t, wrapped.Error(), "30")
}

func TestWrapAPIError_500_PassThrough(t *testing.T) {
	apiErr := makeAPIError(500, "")
	wrapped := WrapAPIError(apiErr, "Port", "uid-123")
	// Non-matched status codes pass through unchanged
	assert.Equal(t, apiErr, wrapped)
}

func TestWrapAPIError_NonAPIError_PassThrough(t *testing.T) {
	plain := errors.New("network timeout")
	wrapped := WrapAPIError(plain, "Port", "uid-123")
	assert.Equal(t, plain, wrapped)
}

func TestWrapAPIError_WrappedErrorUnwrappable(t *testing.T) {
	apiErr := makeAPIError(404, "")
	wrapped := WrapAPIError(apiErr, "VXC", "vxc-789")
	var target *megaport.ErrorResponse
	assert.True(t, errors.As(wrapped, &target))
	assert.Equal(t, 404, target.Response.StatusCode)
}
