package generics

import (
	"fmt"

	"github.com/Kasita-Inc/gadget/errors"
)

// UnsupportedTypeError is returned when an attempt is made to use a value holder with
// a type that is not currently implemented.
type UnsupportedTypeError struct {
	Type  interface{}
	trace []string
}

// NewUnsupportedTypeError for the passed type from a switch statement.
func NewUnsupportedTypeError(obj interface{}) errors.TracerError {
	return &UnsupportedTypeError{
		Type:  obj,
		trace: errors.GetStackTrace(),
	}
}

func (err *UnsupportedTypeError) Error() string {
	return fmt.Sprintf("the type %T is not supported by ValueHolder", err.Type)
}

// Trace returns the stack trace for the error
func (err *UnsupportedTypeError) Trace() []string {
	return err.trace
}

// CorruptValueError is returned when the value contained in a value holder cannot be
// interpreted.
type CorruptValueError struct {
	Value string
	trace []string
}

// NewCorruptValueError is returned when the value inside a value holder cannot be interpreted
func NewCorruptValueError(value string) errors.TracerError {
	return &CorruptValueError{
		Value: value,
		trace: errors.GetStackTrace(),
	}
}

func (err *CorruptValueError) Error() string {
	return fmt.Sprintf("value '%s' cannot be interpreted", err.Value)
}

// Trace returns the stack trace for the error
func (err *CorruptValueError) Trace() []string {
	return err.trace
}
