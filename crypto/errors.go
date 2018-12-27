package crypto

import (
	"github.com/pkg/errors"
)

// ErrNilPrivateKey is raised when a private key was expected but received nil
var ErrNilPrivateKey = errors.New("private key is nil")

// ErrNilPublicKeys is raised when public keys are expected but received nil
var ErrNilPublicKeys = errors.New("public keys are nil")

// ErrNilPublicKey is raised when public key is expected but received nil
var ErrNilPublicKey = errors.New("public key is nil")

// ErrInvalidPublicKeyString is raised when an invalid serialization for a public key is used
var ErrInvalidPublicKeyString = errors.New("invalid public key string")
