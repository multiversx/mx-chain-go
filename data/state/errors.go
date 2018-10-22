package state

import (
	"fmt"
)

// ErrorWrongSize is an error-compatible struct holding 2 values: Expected and Got
type ErrorWrongSize struct {
	Exp int
	Got int
}

// Error returns the error as string
func (e *ErrorWrongSize) Error() string {
	return fmt.Sprintf("wrong size! expected: %d, got %d", e.Exp, e.Got)
}

// NewErrorWrongSize returns a new instantiated struct
func NewErrorWrongSize(exp int, got int) *ErrorWrongSize {
	return &ErrorWrongSize{Exp: exp, Got: got}
}
