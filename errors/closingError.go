package errors

import (
	"errors"
	"strings"
)

// IsClosingError returns true if the provided error is used whenever the node is in the closing process
func IsClosingError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrContextClosing) ||
		errors.Is(err, ErrDBIsClosed) ||
		strings.Contains(err.Error(), ErrDBIsClosed.Error()) ||
		strings.Contains(err.Error(), ErrContextClosing.Error())
}
