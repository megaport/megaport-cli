package validation

import (
	"fmt"
	"strconv"
)

// ParseInt converts a user-supplied string into an int. On failure it returns a
// friendly error naming the field and the offending value, instead of leaking
// strconv internals like `strconv.Atoi: parsing "x": invalid syntax`.
//
// The message stays lowercase ("invalid <field>: ...") so existing exit-code
// classification (which matches a lowercase "invalid"/"ID" substring) and error
// wrapping keep working.
func ParseInt(field, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %q is not a valid whole number", field, value)
	}
	return n, nil
}
