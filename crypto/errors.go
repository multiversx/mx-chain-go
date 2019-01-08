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

// ErrNilHasher is raised when a valid hasher is expected but used nil
var ErrNilHasher = errors.New("marshalizer is nil")

// ErrInvalidIndex is raised when an out of bound index is used
var ErrInvalidIndex = errors.New("index is out of bounds")

// ErrElementNotFound is raised when searching for a speciffic element but found nil
var ErrNilElement = errors.New("element is nil")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil used
var ErrNilKeyGenerator = errors.New("key generator is nil")
