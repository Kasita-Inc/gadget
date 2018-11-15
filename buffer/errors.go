package buffer

import (
	"fmt"

	"github.com/Kasita-Inc/gadget/errors"
)

// TimeoutError is returned when an operation on the connection occurs.
type TimeoutError struct {
	Operation string
	Bytes     int
	trace     []string
}

// NewTimeoutError with the passed operation and bytes that were acted upon prior to the timeout.
func NewTimeoutError(operation string, bytes int) error {
	return &TimeoutError{
		trace:     errors.GetStackTrace(),
		Operation: operation,
		Bytes:     bytes,
	}
}

// Error message string for this error
func (err *TimeoutError) Error() string {
	return fmt.Sprintf("timeout occurred after %s of %d bytes", err.Operation, err.Bytes)
}

// Trace for the error
func (err *TimeoutError) Trace() []string {
	return err.trace
}
