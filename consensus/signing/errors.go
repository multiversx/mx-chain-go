package signing

import "errors"

// ErrInvalidSignature is raised for an invalid signature
var ErrInvalidSignature = errors.New("invalid signature was provided")

// ErrNilElement is raised when searching for a specific element but found nil
var ErrNilElement = errors.New("element is nil")

// ErrIndexNotSelected is raised when a not selected index is used for multi-signing
var ErrIndexNotSelected = errors.New("index is not selected")

// ErrNilBitmap is raised when a nil bitmap is used
var ErrNilBitmap = errors.New("bitmap is nil")

// ErrNoPrivateKeySet is raised when no private key was set
var ErrNoPrivateKeySet = errors.New("no private key was set")

// ErrNoPublicKeySet is raised when no public key was set for a multisignature
var ErrNoPublicKeySet = errors.New("no public key was set")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil used
var ErrNilKeyGenerator = errors.New("key generator is nil")

// ErrNilPublicKeys is raised when public keys are expected but received nil
var ErrNilPublicKeys = errors.New("public keys are nil")

// ErrNilMultiSignerContainer is raised when a nil multi signer container has been provided
var ErrNilMultiSignerContainer = errors.New("multi signer container is nil")

// ErrIndexOutOfBounds is raised when an out of bound index is used
var ErrIndexOutOfBounds = errors.New("index is out of bounds")

// ErrEmptyPubKeyString is raised when an empty public key string is used
var ErrEmptyPubKeyString = errors.New("public key string is empty")

// ErrNilMessage is raised when trying to verify a nil signed message or trying to sign a nil message
var ErrNilMessage = errors.New("message to be signed or to be verified is nil")

// ErrBitmapMismatch is raised when an invalid bitmap is passed to the multisigner
var ErrBitmapMismatch = errors.New("multi signer reported a mismatch in used bitmap")
