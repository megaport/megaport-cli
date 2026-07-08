//go:build !js && !wasm

package auth

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestPrintAuthStatus_XML(t *testing.T) {
	user := &megaport.User{
		FirstName:   "Test",
		LastName:    "User",
		Email:       "test@example.com",
		Position:    "Admin",
		Active:      true,
		CompanyName: "Co",
	}

	out := output.CaptureOutput(func() {
		err := printAuthStatus(user, "profile", "production", "https://api.megaport.com/", "Co", "xml", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<first_name>Test</first_name>")
	assert.Contains(t, out, "<email>test@example.com</email>")
	assert.Contains(t, out, "<environment>Production</environment>")
	assert.Contains(t, out, "<api_endpoint>https://api.megaport.com/</api_endpoint>")
	assert.Contains(t, out, "<active>true</active>")
}

func TestPrintAuthStatus_AllFormatsNoSecretLeak(t *testing.T) {
	user := &megaport.User{FirstName: "Test", Email: "test@example.com"}

	for _, format := range []string{"table", "json", "csv", "xml"} {
		t.Run(format, func(t *testing.T) {
			out := output.CaptureOutput(func() {
				err := printAuthStatus(user, "profile", "production", "https://api.megaport.com/", "Co", format, true)
				assert.NoError(t, err)
			})

			assert.NotContains(t, out, "access_key")
			assert.NotContains(t, out, "secret_key")
			assert.NotContains(t, out, "token")
			assert.NotContains(t, out, "password")
		})
	}
}

func TestPrintAuthStatus_NilUser(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printAuthStatus(nil, "profile", "staging", "https://api-staging.megaport.com/", "Fallback Co", "csv", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "Fallback Co")
	assert.Contains(t, out, "Staging")
}
