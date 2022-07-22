package signing

import "errors"

// ErrNilSignature is raised for a nil signature
var ErrNilSignature = errors.New("signature is nil")

// ErrNilElement is raised when searching for a specific element but found nil
var ErrNilElement = errors.New("element is nil")

// ErrIndexNotSelected is raised when a not selected index is used for multi-signing
var ErrIndexNotSelected = errors.New("index is not selected")

// ErrNilBitmap is raised when a nil bitmap is used
var ErrNilBitmap = errors.New("bitmap is nil")

// ErrNilPrivateKey is raised when a private key was expected but received nil
var ErrNilPrivateKey = errors.New("private key is nil")

// ErrNoPublicKeySet is raised when no public key was set for a multisignature
var ErrNoPublicKeySet = errors.New("no public key was set")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil used
var ErrNilKeyGenerator = errors.New("key generator is nil")

// ErrNilPublicKeys is raised when public keys are expected but received nil
var ErrNilPublicKeys = errors.New("public keys are nil")

// ErrNilSingleSigner singals that a nil single signer has been provided
var ErrNilSingleSigner = errors.New("single signer is nil")

// ErrNilMultiSigner singals that a nil multi signer has been provided
var ErrNilMultiSigner = errors.New("multi signer is nil")

// ErrIndexOutOfBounds is raised when an out of bound index is used
var ErrIndexOutOfBounds = errors.New("index is out of bounds")

// ErrEmptyPubKeyString is raised when an empty public key string is used
var ErrEmptyPubKeyString = errors.New("public key string is empty")

// ErrNilMessage is raised when trying to verify a nil signed message or trying to sign a nil message
var ErrNilMessage = errors.New("message to be signed or to be verified is nil")

// ErrBitmapMismatch is raised when an invalid bitmap is passed to the multisigner
var ErrBitmapMismatch = errors.New("multi signer reported a mismatch in used bitmap")
