//go:build js && wasm

package config

import (
	"context"
	"os"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoginFunc_UnknownEnvironmentFailsClosed verifies that the credential
// login path rejects an unrecognized MEGAPORT_ENVIRONMENT instead of
// silently coercing it to production, since accessKey/secretKey callers
// only reach this switch via setAuthCredentials, which is expected to have
// already bucketed the value into production/staging/development.
func TestLoginFunc_UnknownEnvironmentFailsClosed(t *testing.T) {
	origAccessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	origSecretKey := os.Getenv("MEGAPORT_SECRET_KEY")
	origEnv := os.Getenv("MEGAPORT_ENVIRONMENT")
	defer func() {
		os.Setenv("MEGAPORT_ACCESS_KEY", origAccessKey)
		os.Setenv("MEGAPORT_SECRET_KEY", origSecretKey)
		os.Setenv("MEGAPORT_ENVIRONMENT", origEnv)
	}()

	js.Global().Delete("megaportToken")
	js.Global().Delete("megaportCredentials")

	os.Setenv("MEGAPORT_ACCESS_KEY", "test-access-key")
	os.Setenv("MEGAPORT_SECRET_KEY", "test-secret-key")
	os.Setenv("MEGAPORT_ENVIRONMENT", "not-a-real-environment")

	client, err := loginFunc(context.Background())

	assert.Nil(t, client, "no client should be created for an unrecognized environment")
	assert.ErrorContains(t, err, "unknown environment")
	assert.ErrorContains(t, err, "not-a-real-environment")
}
