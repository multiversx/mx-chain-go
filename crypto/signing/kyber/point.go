package kyber

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"go.dedis.ch/kyber/v3"
)

// kyberPoint is a wrapper over the kyber point interface
type kyberPoint struct {
	kyber.Point
}

// Equal tests if receiver is equal with the Point p given as parameter.
// Both Points need to be derived from the same Group
func (kp *kyberPoint) Equal(p crypto.Point) (bool, error) {
	if check.IfNil(p) {
		return false, crypto.ErrNilParam
	}

	po2, ok := p.(*kyberPoint)

	if !ok {
		return false, crypto.ErrInvalidParam
	}

	eq := kp.Point.Equal(po2.Point)

	return eq, nil
}

// Null returns the neutral identity element.
func (kp *kyberPoint) Null() crypto.Point {
	po2 := kyberPoint{Point: kp.Point.Clone()}
	_ = po2.Point.Null()

	return &po2
}

// Set sets the receiver equal to another Point p.
func (kp *kyberPoint) Set(p crypto.Point) error {
	if p == nil || p.IsInterfaceNil() {
		return crypto.ErrNilParam
	}

	po1, ok := p.(*kyberPoint)

	if !ok {
		return crypto.ErrInvalidParam
	}

	kp.Point.Set(po1.Point)

	return nil
}

// Clone returns a clone of the receiver.
func (kp *kyberPoint) Clone() crypto.Point {
	po2 := kyberPoint{Point: kp.Point.Clone()}

	return &po2
}

// Add returns the result of adding receiver with Point p given as parameter,
// so that their scalars add homomorphically
func (kp *kyberPoint) Add(p crypto.Point) (crypto.Point, error) {
	if p == nil || p.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*kyberPoint)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := kyberPoint{Point: kp.Point.Clone()}
	_ = po2.Point.Add(po2.Point, po1.Point)

	return &po2, nil
}

// Sub returns the result of subtracting from receiver the Point p given as parameter,
// so that their scalars subtract homomorphically
func (kp *kyberPoint) Sub(p crypto.Point) (crypto.Point, error) {
	if p == nil || p.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*kyberPoint)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := kyberPoint{Point: kp.Point.Clone()}
	_ = po2.Point.Sub(po2.Point, po1.Point)

	return &po2, nil
}

// Neg returns the negation of receiver
func (kp *kyberPoint) Neg() crypto.Point {
	po2 := kyberPoint{Point: kp.Point.Clone()}
	_ = po2.Point.Neg(kp.Point)

	return &po2
}

// Mul returns the result of multiplying receiver by the scalar s.
func (kp *kyberPoint) Mul(s crypto.Scalar) (crypto.Point, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s1, ok := s.GetUnderlyingObj().(kyber.Scalar)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := kyberPoint{Point: kp.Point.Clone()}
	_ = po2.Point.Mul(s1, po2.Point)

	return &po2, nil
}

// Pick returns a new random or pseudo-random Point.
func (kp *kyberPoint) Pick(rand cipher.Stream) (crypto.Point, error) {
	if rand == nil {
		return nil, crypto.ErrNilParam
	}

	po2 := kyberPoint{Point: kp.Point.Clone()}
	po2.Point.Pick(rand)

	return &po2, nil
}

// GetUnderlyingObj returns the object the implementation wraps
func (kp *kyberPoint) GetUnderlyingObj() interface{} {
	return kp.Point
}

// MarshalBinary converts the point into its byte array representation
func (kp *kyberPoint) MarshalBinary() ([]byte, error) {
	return kp.Point.MarshalBinary()
}

// UnmarshalBinary reconstructs a point from its byte array representation
func (kp *kyberPoint) UnmarshalBinary(point []byte) error {
	return kp.Point.UnmarshalBinary(point)
}

// IsInterfaceNil returns true if there is no value under the interface
func (kp *kyberPoint) IsInterfaceNil() bool {
	return kp == nil
}
