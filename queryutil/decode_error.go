package queryutil

import "fmt"

type DecodeError struct {
	Field string
	Value string
	Inner error
}

func (de *DecodeError) Error() string {
	return fmt.Sprintf("error parsing field %q value %q: %s", de.Field, de.Value, de.Inner)
}
