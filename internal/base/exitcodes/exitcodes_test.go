package exitcodes

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIError(t *testing.T) {
	t.Run("implements error interface", func(t *testing.T) {
		inner := errors.New("something failed")
		cliErr := New(API, inner)

		var err error = cliErr
		assert.Equal(t, "something failed", err.Error())
	})

	t.Run("Unwrap returns inner error", func(t *testing.T) {
		inner := errors.New("inner")
		cliErr := New(General, inner)
		assert.Equal(t, inner, cliErr.Unwrap())
	})

	t.Run("errors.As extracts CLIError", func(t *testing.T) {
		inner := errors.New("wrapped")
		cliErr := New(Authentication, inner)

		var extracted *CLIError
		require.True(t, errors.As(cliErr, &extracted))
		assert.Equal(t, Authentication, extracted.Code)
		assert.Equal(t, "wrapped", extracted.Error())
	})

	t.Run("errors.As works through wrapping", func(t *testing.T) {
		inner := errors.New("deep")
		cliErr := NewAPIError(inner)

		var extracted *CLIError
		require.True(t, errors.As(cliErr, &extracted))
		assert.Equal(t, API, extracted.Code)
	})
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		wantCode int
	}{
		{"NewUsageError", NewUsageError(errors.New("bad flag")), Usage},
		{"NewAuthError", NewAuthError(errors.New("no creds")), Authentication},
		{"NewAPIError", NewAPIError(errors.New("500")), API},
		{"New with General", New(General, errors.New("unknown")), General},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
		})
	}
}
