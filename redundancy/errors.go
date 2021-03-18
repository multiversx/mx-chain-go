package redundancy

import (
	"errors"
)

// ErrNilMessenger signals that a nil messenger has been provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilObserverPrivateKey signals that a nil observer private key has been provided
var ErrNilObserverPrivateKey = errors.New("nil observer private key")
