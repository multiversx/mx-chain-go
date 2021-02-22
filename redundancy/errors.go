package redundancy

import (
	"errors"
)

// ErrNilMessenger signals that a nil messenger has been provided
var ErrNilMessenger = errors.New("nil messenger")
