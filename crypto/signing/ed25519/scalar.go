package ed25519

import (
	"bytes"
	"crypto/cipher"
	"crypto/ed25519"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

type ed25519Scalar struct {
	ed25519.PrivateKey
}

// Equal checks if the underlying private key inside the scalar objects contain the same bytes
func (es *ed25519Scalar) Equal(s crypto.Scalar) (bool, error) {
	privateKey, err := es.getPrivateKeyFromScalar(s)
	if err != nil {
		return false, err
	}

	return bytes.Equal(privateKey, es.PrivateKey), nil
}

// Set sets the underlying private key inside the scalar to the private key of the provided scalar
func (es *ed25519Scalar) Set(s crypto.Scalar) error {
	privateKey, err := es.getPrivateKeyFromScalar(s)
	if err != nil {
		return err
	}

	es.PrivateKey = privateKey

	return nil
}

// Clone creates a new Scalar with same value as receiver
func (es *ed25519Scalar) Clone() crypto.Scalar {
	scalarBytes := make([]byte, len(es.PrivateKey))
	copy(scalarBytes, es.PrivateKey)

	return &ed25519Scalar{scalarBytes}
}

// GetUnderlyingObj returns the object the implementation wraps
func (es *ed25519Scalar) GetUnderlyingObj() interface{} {
	return es.PrivateKey
}

// MarshalBinary encodes the receiver into a binary form and returns the result.
func (es *ed25519Scalar) MarshalBinary() ([]byte, error) {
	return es.PrivateKey, nil
}

// UnmarshalBinary decodes a scalar from its byte array representation and sets the receiver to this value
func (es *ed25519Scalar) UnmarshalBinary(s []byte) error {
	switch len(s) {
	case ed25519.SeedSize:
		es.PrivateKey = ed25519.NewKeyFromSeed(s)
	case ed25519.PrivateKeySize:
		es.PrivateKey = s
	default:
		return crypto.ErrInvalidPrivateKey
	}

	return nil
}

// SetInt64 is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (es *ed25519Scalar) SetInt64(v int64) {
	panic("ed25519Scalar SetInt64 not implemented")	
}

// Zero is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (es *ed25519Scalar) Zero() crypto.Scalar {
	panic("ed25519Scalar Zero not implemented")
}

// Add is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (es *ed25519Scalar) Add(s crypto.Scalar) (crypto.Scalar, error) {
	panic("ed25519Scalar Add not implemented")
}

// Sub is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (es *ed25519Scalar) Sub(s crypto.Scalar) (crypto.Scalar, error) {
	panic("ed25519Scalar Sub not implemented")
}

// Neg is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (es *ed25519Scalar) Neg() crypto.Scalar {
	panic("ed25519Scalar Neg not implemented")
}

// One is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (es *ed25519Scalar) One() crypto.Scalar {
	panic("ed25519Scalar One not implemented")
}

// Mul returns the modular product of receiver with scalar s given as parameter
func (es *ed25519Scalar) Mul(s crypto.Scalar) (crypto.Scalar, error) {
	panic("ed25519Scalar Mul not implemented")
}

// Div returns the modular division between receiver and scalar s given as parameter
func (es *ed25519Scalar) Div(s crypto.Scalar) (crypto.Scalar, error) {
	panic("ed25519Scalar Div not implemented")
}

// Inv returns the modular inverse of scalar s given as parameter
func (es *ed25519Scalar) Inv(s crypto.Scalar) (crypto.Scalar, error) {
	panic("ed25519Scalar Inv not implemented")
}

// Pick returns a fresh random or pseudo-random scalar
func (es *ed25519Scalar) Pick(rand cipher.Stream) (crypto.Scalar, error) {
	panic("ed25519Scalar Pick not implemented")
}

// SetBytes sets the scalar from a byte-slice,
// reducing if necessary to the appropriate modulus.
func (es *ed25519Scalar) SetBytes(s []byte) (crypto.Scalar, error) {
	panic("ed25519Scalar SetBytes not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (es *ed25519Scalar) IsInterfaceNil() bool {
	return es == nil
}

func(es *ed25519Scalar) getPrivateKeyFromScalar(s crypto.Scalar) (ed25519.PrivateKey, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	privateKey, ok := s.GetUnderlyingObj().(ed25519.PrivateKey)
	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	return privateKey, nil
}
