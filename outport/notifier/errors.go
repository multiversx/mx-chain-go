package notifier

import (
	"errors"
	"fmt"
)

// ErrNilTransactionsPool signals that a nil transactions pool was provided
var ErrNilTransactionsPool = errors.New("nil transactions pool")

const (
	badRequestMessage     = "bad request body"
	unauthorizedMessage   = "unauthorized request"
	internalErrMessage    = "internal server error"
	genericHttpErrMessage = "failed http request"
)

var ErrHttpFailedRequest = func(message string, code int) error {
	return fmt.Errorf("%s, status code = %d", message, code)
}
