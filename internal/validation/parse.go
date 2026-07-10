package validation

import (
	"fmt"
	"strconv"
)

// ParseInt converts a user-supplied string into an int. On failure it returns a
// friendly, typed *ValidationError naming the field and the offending value,
// instead of leaking strconv internals like `strconv.Atoi: parsing "x": invalid
// syntax` or relying on exit-code classification matching the message text.
func ParseInt(field, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		// Quoted so an empty or whitespace-only value renders unambiguously
		// in the error message (e.g. "" rather than a blank gap).
		return 0, NewValidationError(field, fmt.Sprintf("%q", value), "is not a valid whole number")
	}
	return n, nil
}
