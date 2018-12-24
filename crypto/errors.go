package crypto

import (
	"github.com/pkg/errors"
)

// ErrNilPrivateKey is returned when a private key was expected but received nil
var ErrNilPrivateKey = errors.New("private key is nil")

// ErrNilPublicKeys is returned when public keys expected but received nil
var ErrNilPublicKeys = errors.New("public keys are nil")
