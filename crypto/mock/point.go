package mock

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

type PointMock struct {
	X int
	Y int

	MarshalBinaryStub   func(x, y int) ([]byte, error)
	UnmarshalBinaryStub func([]byte) (x, y int, err error)
}

// Equal tests if receiver is equal with the Point p given as parameter.
// Both Points need to be derived from the same Group
func (po *PointMock) Equal(p crypto.Point) (bool, error) {
	pp, ok := p.(*PointMock)

	if !ok {
		return false, crypto.ErrInvalidParam
	}

	return pp.X == po.X && pp.Y == po.Y, nil
}

// Null returns the neutral identity element.
func (po *PointMock) Null() crypto.Point {
	panic("implement me")
}

// Base returns the Group's base point.
func (po *PointMock) Base() crypto.Point {
	pp := &PointMock{
		X:                   1,
		Y:                   1,
		UnmarshalBinaryStub: po.UnmarshalBinaryStub,
		MarshalBinaryStub:   po.MarshalBinaryStub,
	}

	return pp
}

// Set sets the receiver equal to another Point p.
func (po *PointMock) Set(p crypto.Point) error {
	panic("implement me")
}

// Clone returns a clone of the receiver.
func (po *PointMock) Clone() crypto.Point {
	panic("implement me")
}

// Add returns the result of adding receiver with Point p given as parameter,
// so that their scalars add homomorphically
func (po *PointMock) Add(p crypto.Point) (crypto.Point, error) {
	panic("implement me")
}

// Sub returns the result of subtracting from receiver the Point p given as parameter,
// so that their scalars subtract homomorphically
func (po *PointMock) Sub(p crypto.Point) (crypto.Point, error) {
	panic("implement me")
}

// Neg returns the negation of receiver
func (po *PointMock) Neg() crypto.Point {
	panic("implement me")
}

// Mul returns the result of multiplying receiver by the scalar s.
// Mock multiplies the scalar to both X and Y fields
func (po *PointMock) Mul(s crypto.Scalar) (crypto.Point, error) {
	if s == nil {
		return nil, crypto.ErrInvalidParam
	}

	ss, _ := s.(*ScalarMock)

	pp := &PointMock{
		X:                   po.X * ss.X,
		Y:                   po.Y * ss.X,
		UnmarshalBinaryStub: po.UnmarshalBinaryStub,
		MarshalBinaryStub:   po.MarshalBinaryStub,
	}

	return pp, nil
}

// Pick returns a fresh random or pseudo-random Point.
func (po *PointMock) Pick(rand cipher.Stream) (crypto.Point, error) {
	panic("implement me")
}

// GetUnderlyingObj returns the object the implementation wraps
func (po *PointMock) GetUnderlyingObj() interface{} {
	return 0
}

// MarshalBinary transforms the Point into a byte array
func (po *PointMock) MarshalBinary() ([]byte, error) {
	return po.MarshalBinaryStub(po.X, po.Y)
}

// UnmarshalBinary recreates the Point from a byte array
func (po *PointMock) UnmarshalBinary(point []byte) error {
	x, y, err := po.UnmarshalBinaryStub(point)
	po.X = x
	po.Y = y

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (po *PointMock) IsInterfaceNil() bool {
	if po == nil {
		return true
	}
	return false
}
