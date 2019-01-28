package kv2

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"gopkg.in/dedis/kyber.v2"
)

// kyberPoint is a wrapper over the kyber point interface
type kyberPoint struct {
	kyber.Point
}

// Equal tests if receiver is equal with the Point p given as parameter.
// Both Points need to be derived from the same Group
func (po *kyberPoint) Equal(p crypto.Point) (bool, error) {
	if p == nil {
		return false, crypto.ErrNilParam
	}

	po2, ok := p.(*kyberPoint)

	if !ok {
		return false, crypto.ErrInvalidParam
	}

	eq := po.Point.Equal(po2.Point)

	return eq, nil
}

// Null returns the neutral identity element.
func (po *kyberPoint) Null() crypto.Point {
	po2 := kyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Null()

	return &po2
}

// Base returns the Group's base point.
func (po *kyberPoint) Base() crypto.Point {
	po2 := kyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Base()

	return &po2
}

// Set sets the receiver equal to another Point p.
func (po *kyberPoint) Set(p crypto.Point) error {
	if p == nil {
		return crypto.ErrNilParam
	}

	po1, ok := p.(*kyberPoint)

	if !ok {
		return crypto.ErrInvalidParam
	}

	po.Point.Set(po1.Point)

	return nil
}

// Clone returns a clone of the receiver.
func (po *kyberPoint) Clone() crypto.Point {
	po2 := kyberPoint{Point: po.Point.Clone()}

	return &po2
}

// Data returns the Point as a byte array.
// Returns an error if doesn't represent valid point.
func (po *kyberPoint) Data() ([]byte, error) {
	data, err := po.Point.Data()

	return data, err
}

// Embedd sets the receiver using a Point byte array representation
// This is the inverse operation of Data
func (po *kyberPoint) Embed(data []byte, rand cipher.Stream) error {
	if data == nil || rand == nil {
		return crypto.ErrNilParam
	}

	_ = po.Point.Embed(data, rand)

	return nil
}

// Add returns the result of adding receiver with Point p given as parameter,
// so that their scalars add homomorphically
func (po *kyberPoint) Add(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*kyberPoint)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := kyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Add(po2.Point, po1.Point)

	return &po2, nil
}

// Sub returns the result of subtracting from receiver the Point p given as parameter,
// so that their scalars subtract homomorphically
func (po *kyberPoint) Sub(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*kyberPoint)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := kyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Sub(po2.Point, po1.Point)

	return &po2, nil
}

// Neg returns the negation of receiver
func (po *kyberPoint) Neg() crypto.Point {
	po2 := kyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Neg(po.Point)

	return &po2
}

// Mul returns the result of multiplying receiver by the scalar s.
func (po *kyberPoint) Mul(s crypto.Scalar) (crypto.Point, error) {
	if s == nil {
		return nil, crypto.ErrNilParam
	}

	s1, ok := s.GetUnderlyingObj().(kyber.Scalar)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := kyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Mul(s1, po2.Point)

	return &po2, nil
}

// Pick returns a new random or pseudo-random Point.
func (po *kyberPoint) Pick(rand cipher.Stream) (crypto.Point, error) {
	if rand == nil {
		return nil, crypto.ErrNilParam
	}

	po2 := kyberPoint{Point: po.Point.Clone()}
	po2.Point.Pick(rand)

	return &po2, nil
}

// GetUnderlyingObj returns the object the implementation wraps
func (po *kyberPoint) GetUnderlyingObj() interface{} {
	return po.Point
}

// MarshalBinary converts the point into its byte array representation
func (po *kyberPoint) MarshalBinary() ([]byte, error) {
	return po.Point.MarshalBinary()
}

// UnmarshalBinary reconstructs a point from its byte array representation
func (po *kyberPoint) UnmarshalBinary(point []byte) error {
	return po.Point.UnmarshalBinary(point)
}
