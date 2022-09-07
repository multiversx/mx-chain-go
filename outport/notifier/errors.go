package notifier

import (
	"errors"
)

// ErrNilTransactionsPool signals that a nil transactions pool was provided
var ErrNilTransactionsPool = errors.New("nil transactions pool")
