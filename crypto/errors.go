package crypto

import (
	"github.com/pkg/errors"
)

// ErrNilPrivateKey is raised when a private key was expected but received nil
var ErrNilPrivateKey = errors.New("private key is nil")

// ErrInvalidPrivateKey is raised when an invalid private key is used
var ErrInvalidPrivateKey = errors.New("private key is invalid")

// ErrNilPublicKeys is raised when public keys are expected but received nil
var ErrNilPublicKeys = errors.New("public keys are nil")

// ErrNilPublicKey is raised when public key is expected but received nil
var ErrNilPublicKey = errors.New("public key is nil")

// ErrInvalidPublicKey is raised when an invalid public key is used
var ErrInvalidPublicKey = errors.New("public key is invalid")

// ErrInvalidPublicKeyString is raised when an invalid serialization for a public key is used
var ErrInvalidPublicKeyString = errors.New("invalid public key string")

// ErrNilHasher is raised when a valid hasher is expected but used nil
var ErrNilHasher = errors.New("marshalizer is nil")

// ErrInvalidIndex is raised when an out of bound index is used
var ErrInvalidIndex = errors.New("index is out of bounds")

// ErrNilElement is raised when searching for a specific element but found nil
var ErrNilElement = errors.New("element is nil")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil used
var ErrNilKeyGenerator = errors.New("key generator is nil")

// ErrNilParam is raised for nil parameters
var ErrNilParam = errors.New("nil parameter")

// ErrInvalidParam is raised for invalid parameters
var ErrInvalidParam = errors.New("parameter is invalid")

// ErrNilSuite is raised when a nil crypto suite is used
var ErrNilSuite = errors.New("crypto suite is nil")

// ErrInvalidSuite is raised when an invalid crypto suite is used
var ErrInvalidSuite = errors.New("crypto suite is invalid")

// ErrNilSignature is raised for a nil signature
var ErrNilSignature = errors.New("signature is nil")

// ErrNilMessage is raised when trying to verify a nil signed message or trying to sign a nil message
var ErrNilMessage = errors.New("message to be signed or to be verified is nil")

// ErrNilSingleSigner is raised when using a nil single signer
var ErrNilSingleSigner = errors.New("single signer is nil")
