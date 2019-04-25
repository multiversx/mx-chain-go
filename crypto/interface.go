package crypto

import (
	"crypto/cipher"
)

// A Scalar represents a scalar value by which
// a Point (group element) may be encrypted to produce another Point.
// adapted from kyber
type Scalar interface {
	// MarshalBinary transforms the Scalar into a byte array
	MarshalBinary() ([]byte, error)
	// UnmarshalBinary recreates the Scalar from a byte array
	UnmarshalBinary([]byte) error
	// Equal tests if receiver is equal with the scalar s given as parameter.
	// Both scalars need to be derived from the same Group
	Equal(s Scalar) (bool, error)
	// Set sets the receiver to Scalar s given as parameter
	Set(s Scalar) error
	// Clone creates a new Scalar with same value as receiver
	Clone() Scalar
	// SetInt64 sets the receiver to a small integer value v given as parameter
	SetInt64(v int64)
	// Zero returns the the additive identity (0)
	Zero() Scalar
	// Add returns the modular sum of receiver with scalar s given as parameter
	Add(s Scalar) (Scalar, error)
	// Sub returns the modular difference between receiver and scalar s given as parameter
	Sub(s Scalar) (Scalar, error)
	// Neg returns the modular negation of receiver
	Neg() Scalar
	// One returns the multiplicative identity (1)
	One() Scalar
	// Mul returns the modular product of receiver with scalar s given as parameter
	Mul(s Scalar) (Scalar, error)
	// Div returns the modular division between receiver and scalar s given as parameter
	Div(s Scalar) (Scalar, error)
	// Inv returns the modular inverse of scalar s given as parameter
	Inv(s Scalar) (Scalar, error)
	// Pick returns a fresh random or pseudo-random scalar
	Pick(rand cipher.Stream) (Scalar, error)
	// SetBytes sets the scalar from a byte-slice,
	// reducing if necessary to the appropriate modulus.
	SetBytes([]byte) (Scalar, error)
	// GetUnderlyingObj returns the object the implementation wraps
	GetUnderlyingObj() interface{}
}

// Point represents an element of a public-key cryptographic Group.
// adapted from kyber
type Point interface {
	// MarshalBinary transforms the Point into a byte array
	MarshalBinary() ([]byte, error)
	// UnmarshalBinary recreates the Point from a byte array
	UnmarshalBinary([]byte) error
	// Equal tests if receiver is equal with the Point p given as parameter.
	// Both Points need to be derived from the same Group
	Equal(p Point) (bool, error)
	// Null returns the neutral identity element.
	Null() Point
	// Base returns the Group's base point.
	Base() Point
	// Set sets the receiver equal to another Point p.
	Set(p Point) error
	// Clone returns a clone of the receiver.
	Clone() Point
	// Add returns the result of adding receiver with Point p given as parameter,
	// so that their scalars add homomorphically
	Add(p Point) (Point, error)
	// Sub returns the result of subtracting from receiver the Point p given as parameter,
	// so that their scalars subtract homomorphically
	Sub(p Point) (Point, error)
	// Neg returns the negation of receiver
	Neg() Point
	// Mul returns the result of multiplying receiver by the scalar s.
	Mul(s Scalar) (Point, error)
	// Pick returns a fresh random or pseudo-random Point.
	Pick(rand cipher.Stream) (Point, error)
	// GetUnderlyingObj returns the object the implementation wraps
	GetUnderlyingObj() interface{}
}

// Group defines a mathematical group used for Diffie-Hellmann operations
// adapted from kyber
type Group interface {
	// String returns the string for the group
	String() string
	// ScalarLen returns the maximum length of scalars in bytes
	ScalarLen() int
	// CreateScalar creates a new Scalar
	CreateScalar() Scalar
	// PointLen returns the max length of point in nb of bytes
	PointLen() int
	// CreatePoint creates a new point
	CreatePoint() Point
}

// Random is an interface that can be mixed in to local suite definitions.
// adapted from kyber
type Random interface {
	// RandomStream returns a cipher.Stream that produces a
	// cryptographically random key stream. The stream must
	// tolerate being used in multiple goroutines.
	RandomStream() cipher.Stream
}

// Suite represents the list of functionalities needed by this package.
// adapted from kyber
type Suite interface {
	Group
	Random
	CreateKeyPair(cipher.Stream) (Scalar, Point)
	GetUnderlyingSuite() interface{}
}

// KeyGenerator is an interface for generating different types of cryptographic keys
type KeyGenerator interface {
	GeneratePair() (PrivateKey, PublicKey)
	PrivateKeyFromByteArray(b []byte) (PrivateKey, error)
	PublicKeyFromByteArray(b []byte) (PublicKey, error)
	Suite() Suite
}

// Key represents a crypto key - can be either private or public
type Key interface {
	// ToByteArray returns the byte array representation of the key
	ToByteArray() ([]byte, error)
	// Suite returns the suite used by this key
	Suite() Suite
}

// PrivateKey represents a private key that can sign data or decrypt messages encrypted with a public key
type PrivateKey interface {
	Key
	// GeneratePublic builds a public key for the current private key
	GeneratePublic() PublicKey
	// Scalar returns the Scalar corresponding to this Private Key
	Scalar() Scalar
}

// PublicKey can be used to encrypt messages
type PublicKey interface {
	Key
	// Point returns the Point corresponding to this Public Key
	Point() Point
}

// SingleSigner provides functionality for signing a message and verifying a single signed message
type SingleSigner interface {
	// Sign is used to sign a message
	Sign(private PrivateKey, msg []byte) ([]byte, error)
	// Verify is used to verify a signed message
	Verify(public PublicKey, msg []byte, sig []byte) error
}

// MultiSigner provides functionality for multi-signing a message and verifying a multi-signed message
type MultiSigner interface {
	// Reset resets the data holder inside the multiSigner
	Reset(pubKeys []string, index uint16) error
	// MultiSigVerifier Provides functionality for verifying a multi-signature
	MultiSigVerifier
	// CreateCommitment creates a secret commitment and the corresponding public commitment point
	CreateCommitment() (commSecret []byte, commitment []byte)
	// StoreCommitmentHash adds a commitment hash to the list with the specified position
	StoreCommitmentHash(index uint16, commHash []byte) error
	// CommitmentHash returns the commitment hash from the list with the specified position
	CommitmentHash(index uint16) ([]byte, error)
	// StoreCommitment adds a commitment to the list with the specified position
	StoreCommitment(index uint16, value []byte) error
	// Commitment returns the commitment from the list with the specified position
	Commitment(index uint16) ([]byte, error)
	// AggregateCommitments aggregates the list of commitments
	AggregateCommitments(bitmap []byte) error
	// CreateSignatureShare creates a partial signature
	CreateSignatureShare(bitmap []byte) ([]byte, error)
	// StoreSignatureShare adds the partial signature of the signer with specified position
	StoreSignatureShare(index uint16, sig []byte) error
	// SignatureShare returns the partial signature set for given index
	SignatureShare(index uint16) ([]byte, error)
	// VerifySignatureShare verifies the partial signature of the signer with specified position
	VerifySignatureShare(index uint16, sig []byte, bitmap []byte) error
	// AggregateSigs aggregates all collected partial signatures
	AggregateSigs(bitmap []byte) ([]byte, error)
}

// MultiSigVerifier Provides functionality for verifying a multi-signature
type MultiSigVerifier interface {
	// Create resets the multisigner and initializes to the new params
	Create(pubKeys []string, index uint16) (MultiSigner, error)
	// SetMessage sets the message to be multi-signed upon
	SetMessage(msg []byte) error
	// SetAggregatedSig sets the aggregated signature
	SetAggregatedSig([]byte) error
	// Verify verifies the aggregated signature
	Verify(bitmap []byte) error
}

// MultiSignerBLS provides functionality for multi-signing a message and ferifying a multi-signed message
// TODO: refactor BN to use this same multiSigner, and then remove the MultiSigner - EN-1774
type MultiSignerBLS interface {
	// MultiSigVerifierBLS Provides functionality for verifying a multi-signature
	MultiSigVerifierBLS
	// Reset resets the data holder inside the multiSigner
	Reset(pubKeys []string, index uint16) error
	// CreateSignatureShare creates a partial signature
	CreateSignatureShare() ([]byte, error)
	// StoreSignatureShare adds the partial signature of the signer with specified position
	StoreSignatureShare(index uint16, sig []byte) error
	// SignatureShare returns the partial signature set for given index
	SignatureShare(index uint16) ([]byte, error)
	// VerifySignatureShare verifies the partial signature of the signer with specified position
	VerifySignatureShare(index uint16, sig []byte) error
	// AggregateSigs aggregates all collected partial signatures
	AggregateSigs(bitmap []byte) ([]byte, error)
}

// MultiSigVerifierBLS Provides functionality for verifying a multi-signature
type MultiSigVerifierBLS interface {
	// Create resets the multisigner and initializes to the new params
	Create(pubKeys []string, index uint16) (MultiSignerBLS, error)
	// SetMessage sets the message to be multi-signed upon
	SetMessage(msg []byte) error
	// SetAggregatedSig sets the aggregated signature
	SetAggregatedSig([]byte) error
	// Verify verifies the aggregated signature
	Verify(bitmap []byte) error
}

type LowLevelSignerBLS interface {
	VerifySigShare(pubKey PublicKey, message []byte, sig []byte) error
	SignShare(privKey PrivateKey, message []byte) ([]byte, error)
	VerifySigBytes(suite Suite, sig []byte) error
	AggregateSignatures(suite Suite, sigs ...[]byte) ([]byte, error)
	VerifyAggregatedSig(suite Suite, aggPointsBytes []byte, aggSigBytes []byte, msg []byte) error
	AggregatePublicKeys(suite Suite, pubKeys ...Point) ([]byte, error)
	ScalarMulSig(suite Suite, scalar Scalar, sig []byte) ([]byte, error)
}
