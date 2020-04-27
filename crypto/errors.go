package crypto

import (
	"errors"
)

// ErrNilPrivateKey is raised when a private key was expected but received nil
var ErrNilPrivateKey = errors.New("private key is nil")

// ErrInvalidPrivateKey is raised when an invalid private key is used
var ErrInvalidPrivateKey = errors.New("private key is invalid")

// ErrNilPrivateKeyScalar is raised when a private key with nil scalar is used
var ErrNilPrivateKeyScalar = errors.New("private key holds a nil scalar")

// ErrInvalidScalar is raised when an invalid scalar is used
var ErrInvalidScalar = errors.New("scalar is invalid")

// ErrInvalidPoint is raised when an invalid point is used
var ErrInvalidPoint = errors.New("point is invalid")

// ErrNilPublicKeys is raised when public keys are expected but received nil
var ErrNilPublicKeys = errors.New("public keys are nil")

// ErrNilPublicKey is raised when public key is expected but received nil
var ErrNilPublicKey = errors.New("public key is nil")

// ErrInvalidPublicKey is raised when an invalid public key is used
var ErrInvalidPublicKey = errors.New("public key is invalid")

// ErrNilPublicKeyPoint is raised when a public key with nil point is used
var ErrNilPublicKeyPoint = errors.New("public key holds a nil point")

// ErrNoPublicKeySet is raised when no public key was set for a multisignature
var ErrNoPublicKeySet = errors.New("no public key was set")

// ErrInvalidPublicKeyString is raised when an invalid serialization for a public key is used
var ErrInvalidPublicKeyString = errors.New("invalid public key string")

// ErrNilHasher is raised when a valid hasher is expected but used nil
var ErrNilHasher = errors.New("hasher is nil")

// ErrWrongSizeHasher is raised when a hasher with a wrong output size is used
var ErrWrongSizeHasher = errors.New("wrong size hasher")

// ErrIndexOutOfBounds is raised when an out of bound index is used
var ErrIndexOutOfBounds = errors.New("index is out of bounds")

// ErrIndexNotSelected is raised when a not selected index is used for multi-signing
var ErrIndexNotSelected = errors.New("index is not selected")

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

// ErrNilSignaturesList is raised when a nil list of signatures is provided
var ErrNilSignaturesList = errors.New("signature list is nil")

// ErrNilMessage is raised when trying to verify a nil signed message or trying to sign a nil message
var ErrNilMessage = errors.New("message to be signed or to be verified is nil")

// ErrBitmapMismatch is raised when an invalid bitmap is passed to the multisigner
var ErrBitmapMismatch = errors.New("multi signer reported a mismatch in used bitmap")

// ErrNilBitmap is raised when a nil bitmap is used
var ErrNilBitmap = errors.New("bitmap is nil")

// ErrSigNotValid is raised when a signature verification fails due to invalid signature
var ErrSigNotValid = errors.New("signature is invalid")

// ErrAggSigNotValid is raised when an aggregate signature is invalid
var ErrAggSigNotValid = errors.New("aggregate signature is invalid")

// ErrEmptyPubKeyString is raised when an empty public key string is used
var ErrEmptyPubKeyString = errors.New("public key string is empty")

// ErrInvalidSigner is raised when the signer is invalid
var ErrInvalidSigner = errors.New("signer is invalid")

// ErrEd25519InvalidSignature will be returned when ed25519 signature verification fails
var ErrEd25519InvalidSignature = errors.New("ed25519: invalid signature")

// ErrBLSInvalidSignature will be returned when the provided BLS signature is invalid
var ErrBLSInvalidSignature = errors.New("bls12-381: invalid signature")

// ErrGeneratingPubFromPriv signals that there was an error generating a public key corresponding to a provided private key
var ErrGeneratingPubFromPriv = errors.New("unable to generate PublicKey from provided private key")

// ErrNotImplemented signals that a method is not implemented for an interface implementation
var ErrNotImplemented = errors.New("not implemented")

// ErrWrongPrivateKeySize signals that the length of the provided private key is not the expected one
var ErrWrongPrivateKeySize = errors.New("wrong private key size")

// ErrWrongPrivateKeyStructure signals that the structure of the private key is incorrect
var ErrWrongPrivateKeyStructure = errors.New("wrong private key structure")
