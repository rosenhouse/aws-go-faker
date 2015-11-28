package queryutil

import "fmt"

type decodeError struct {
	Field string
	Value string
	Inner error
}

func (de *decodeError) Error() string {
	return fmt.Sprintf("error parsing field %q value %q: %s", de.Field, de.Value, de.Inner)
}
