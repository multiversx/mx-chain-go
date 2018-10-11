package state

import (
	"fmt"
)

type ErrorWrongSize struct {
	Exp int
	Got int
}

func (e *ErrorWrongSize) Error() string {
	return fmt.Sprintf("wrong size! expected: %d, got %d", e.Exp, e.Got)
}

func NewErrorWrongSize(exp int, got int) *ErrorWrongSize {
	return &ErrorWrongSize{Exp: exp, Got: got}
}
