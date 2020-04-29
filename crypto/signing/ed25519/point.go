package ed25519

import (
	"bytes"
	"crypto/ed25519"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

var _ crypto.Point = (*ed25519Point)(nil)

// ed25519 - is a mapping over crypto/ed25519 public key
// Most of the implementations of the Point interface functions are mocked
//  since the implementation details edwards25519 are internal to the crypto/ed25519 package
type ed25519Point struct {
	ed25519.PublicKey
}

// Equal tests if receiver is equal with the Point p given as parameter.
func (ep *ed25519Point) Equal(p crypto.Point) (bool, error) {
	if check.IfNil(p) {
		return false, crypto.ErrNilParam
	}

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
	if check.IfNil(p) {
		return crypto.ErrNilParam
	}

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
func (ep *ed25519Point) Null() crypto.Point {
	log.Error("ed25519Point",
		"message", "Null for ed25519Point is not implemented, should not be called")

	return nil
}

// Base is not needed for this use case, should be removed if possible
func (ep *ed25519Point) Base() crypto.Point {
	log.Error("ed25519Point",
		"message", "Base for ed25519Point is not implemented, should not be called")

	return nil
}

// Add is not needed for this use case, should be removed if possible
func (ep *ed25519Point) Add(_ crypto.Point) (crypto.Point, error) {
	log.Error("ed25519Point",
		"message", "Add for ed25519Point is not implemented, should not be called")

	return nil, crypto.ErrNotImplemented
}

// Sub is not needed for this use case, should be removed if possible
func (ep *ed25519Point) Sub(_ crypto.Point) (crypto.Point, error) {
	log.Error("ed25519Point",
		"message", "Sub for ed25519Point is not implemented, should not be called")

	return nil, crypto.ErrNotImplemented
}

// Neg is not needed for this use case, should be removed if possible
func (ep *ed25519Point) Neg() crypto.Point {
	log.Error("ed25519Point",
		"message", "Neg for ed25519Point is not implemented, should not be called")

	return nil
}

// Mul is not needed for this use case, should be removed if possible
func (ep *ed25519Point) Mul(_ crypto.Scalar) (crypto.Point, error) {
	log.Error("ed25519Point",
		"message", "Mul for ed25519Point is not implemented, should not be called")

	return nil, crypto.ErrNotImplemented
}

// Pick is not needed for this use case, should be removed if possible
func (ep *ed25519Point) Pick() (crypto.Point, error) {
	log.Error("ed25519Point",
		"message", "Pick for ed25519Point is not implemented, should not be called")

	return nil, crypto.ErrNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (ep *ed25519Point) IsInterfaceNil() bool {
	return ep == nil
}
