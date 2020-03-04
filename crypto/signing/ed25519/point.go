package ed25519

import (
	"bytes"
	"crypto/cipher"
	"crypto/ed25519"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

// ed25519 - is a mapping over crypto/ed25519 public key
// Most of the implementations of the Point interface functions are mocked
//  since the implementation details edwards25519 are internal to the crypto/ed25519 package
type ed25519Point struct {
	ed25519.PublicKey
}

// Equal tests if receiver is equal with the Point p given as parameter.
func (ep *ed25519Point) Equal(p crypto.Point) (bool, error) {
	ed25519P, ok := p.(*ed25519Point)
	if !ok {
		return false, crypto.ErrInvalidPublicKey
	}

	return bytes.Equal(ep.PublicKey, ed25519P.PublicKey), nil
}
// GetUnderlyingObj returns the object the implementation wraps
func (ep *ed25519Point) GetUnderlyingObj() interface{} {
	return ep.PublicKey
}

// MarshalBinary converts the point into its byte array representation
func (ep *ed25519Point) MarshalBinary() ([]byte, error) {
	return ep.PublicKey, nil
}

// UnmarshalBinary reconstructs a point from its byte array representation
func (ep *ed25519Point) UnmarshalBinary(point []byte) error {
	ep.PublicKey = point
	return nil
}

// Set sets the receiver equal to another Point p.
func (ep *ed25519Point) Set(p crypto.Point) error {
	point, ok := p.(*ed25519Point)
	if !ok {
		return crypto.ErrInvalidPublicKey
	}

	ep.PublicKey = point.PublicKey
	return nil
}

// Clone returns a clone of the receiver.
func (ep *ed25519Point) Clone() crypto.Point {
	publicKeyBytes := make([]byte, len(ep.PublicKey))
	copy(publicKeyBytes, ep.PublicKey)

	return &ed25519Point{publicKeyBytes}
}

// Null is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (ep *ed25519Point) Null() crypto.Point {
	panic("ed25519Point Null not implemented")
}

// Base is not needed for this use case, should be removed if possible
// - panics to prevent using it
func (ep *ed25519Point) Base() crypto.Point {
	panic("ed25519Point Base not implemented")
}

// Add is not needed for this use case, should be removed if possible
// - panics to prevent using it
func (ep *ed25519Point) Add(p crypto.Point) (crypto.Point, error) {
	panic("ed25519Point Add not implemented")
}

// Sub is not needed for this use case, should be removed if possible
// - panics to prevent using it
func (ep *ed25519Point) Sub(p crypto.Point) (crypto.Point, error) {
	panic("ed25519Point Sub not implemented")
}

// Neg is not needed for this use case, should be removed if possible
// - panics to prevent using it
func (ep *ed25519Point) Neg() crypto.Point {
	panic("ed25519Point Neg not implemented")
}

// Mul is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (ep *ed25519Point) Mul(s crypto.Scalar) (crypto.Point, error) {
	panic("ed25519Point Mul not implemented")
}

// Pick is not needed for this use case, should be removed if possible
//  - panics to prevent using it
func (ep *ed25519Point) Pick(rand cipher.Stream) (crypto.Point, error) {
	panic("ed25519Point Pick not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ep *ed25519Point) IsInterfaceNil() bool {
	return ep == nil
}
