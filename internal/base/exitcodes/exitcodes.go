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
