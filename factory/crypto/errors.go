package crypto

import "errors"

// ErrInvalidSignature is raised for an invalid signature
var ErrInvalidSignature = errors.New("invalid signature was provided")

// ErrNilElement is raised when searching for a specific element but found nil
var ErrNilElement = errors.New("element is nil")

// ErrNilBitmap is raised when a nil bitmap is used
var ErrNilBitmap = errors.New("bitmap is nil")

// ErrNilKeysHandler is raised when a nil keys handler was provided
var ErrNilKeysHandler = errors.New("nil keys handler")

// ErrNoPublicKeySet is raised when no public key was set for a multisignature
var ErrNoPublicKeySet = errors.New("no public key was set")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil used
var ErrNilKeyGenerator = errors.New("key generator is nil")

// ErrNilPublicKeys is raised when public keys are expected but received nil
var ErrNilPublicKeys = errors.New("public keys are nil")

// ErrNilMultiSignerContainer is raised when a nil multi signer container has been provided
var ErrNilMultiSignerContainer = errors.New("multi signer container is nil")

// ErrNilSingleSigner signals that a nil single signer was provided
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrIndexOutOfBounds is raised when an out of bound index is used
var ErrIndexOutOfBounds = errors.New("index is out of bounds")

// ErrEmptyPubKeyString is raised when an empty public key string is used
var ErrEmptyPubKeyString = errors.New("public key string is empty")

// ErrNilMessage is raised when trying to verify a nil signed message or trying to sign a nil message
var ErrNilMessage = errors.New("message to be signed or to be verified is nil")

// ErrBitmapMismatch is raised when an invalid bitmap is passed to the multisigner
var ErrBitmapMismatch = errors.New("multi signer reported a mismatch in used bitmap")
