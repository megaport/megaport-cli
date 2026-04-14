package exitcodes

const (
	Success        = 0
	General        = 1
	Usage          = 2
	Authentication = 3
	API            = 4
	Cancelled      = 5
)

// CLIError wraps an error with a specific exit code.
type CLIError struct {
	Code int
	Err  error
}

func (e *CLIError) Error() string { return e.Err.Error() }
func (e *CLIError) Unwrap() error { return e.Err }

func New(code int, err error) *CLIError     { return &CLIError{Code: code, Err: err} }
func NewUsageError(err error) *CLIError     { return &CLIError{Code: Usage, Err: err} }
func NewAuthError(err error) *CLIError      { return &CLIError{Code: Authentication, Err: err} }
func NewAPIError(err error) *CLIError       { return &CLIError{Code: API, Err: err} }
func NewCancelledError(err error) *CLIError { return &CLIError{Code: Cancelled, Err: err} }

// TypeName returns the string error type name for a given exit code.
// Used when emitting structured JSON error output (--output json).
func TypeName(code int) string {
	switch code {
	case Usage:
		return "usage_error"
	case Authentication:
		return "auth_error"
	case API:
		return "api_error"
	case Cancelled:
		return "cancelled"
	default:
		return "general_error"
	}
}
