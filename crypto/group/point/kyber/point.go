package kyber

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/group"
	scalar "github.com/ElrondNetwork/elrond-go-sandbox/crypto/group/scalar/kyber"
	"gopkg.in/dedis/kyber.v2"
)

// KyberPoint is a wrapper over the kyber point interface
type KyberPoint struct {
	kyber.Point
}

// Equal tests if receiver is equal with the Point p given as parameter.
// Both Points need to be derived from the same Group
func (po *KyberPoint) Equal(p crypto.Point) (bool, error) {
	if p == nil {
		return false, group.ErrNilParam
	}

	po2, ok := p.(*KyberPoint)

	if !ok {
		return false, group.ErrInvalidParam
	}

	eq := po.Point.Equal(po2.Point)

	return eq, nil
}

// Null returns the neutral identity element.
func (po *KyberPoint) Null() crypto.Point {
	po2 := KyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Null()

	return &po2
}

// Base returns the Group's base point.
func (po *KyberPoint) Base() crypto.Point {
	po2 := KyberPoint{Point: po.Point.Clone()}
	_ = po2.Base()

	return &po2
}

// Set sets the receiver equal to another Point p.
func (po *KyberPoint) Set(p crypto.Point) error {
	if p == nil {
		return group.ErrNilParam
	}

	po1, ok := p.(*KyberPoint)

	if !ok {
		return group.ErrInvalidParam
	}

	po.Point.Set(po1.Point)

	return nil
}

// Clone returns a clone of the receiver.
func (po *KyberPoint) Clone() crypto.Point {
	po2 := KyberPoint{Point: po.Point.Clone()}

	return &po2
}

// Data returns the Point as a byte array.
// Returns an error if doesn't represent valid point.
func (po *KyberPoint) Data() ([]byte, error) {
	data, err := po.Point.Data()

	return data, err
}

// Embedd sets the receiver using a Point byte array representation
// This is the inverse operation of Data
func (po *KyberPoint) Embed(data []byte, rand cipher.Stream) error {
	if data == nil || rand == nil {
		return group.ErrNilParam
	}

	_ = po.Point.Embed(data, rand)

	return nil
}

// Add returns the result of adding receiver with Point p given as parameter,
// so that their scalars add homomorphically
func (po *KyberPoint) Add(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, group.ErrNilParam
	}

	po1, ok := p.(*KyberPoint)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	po2 := KyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Add(po2.Point, po1.Point)

	return &po2, nil
}

// Sub returns the result of subtracting from receiver the Point p given as parameter,
// so that their scalars subtract homomorphically
func (po *KyberPoint) Sub(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, group.ErrNilParam
	}

	po1, ok := p.(*KyberPoint)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	po2 := KyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Sub(po2.Point, po1.Point)

	return &po2, nil
}

// Neg returns the negation of receiver
func (po *KyberPoint) Neg() crypto.Point {
	po2 := KyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Neg(po)

	return &po2
}

// Mul returns the result of multiplying receiver by the scalar s.
func (po *KyberPoint) Mul(s crypto.Scalar) (crypto.Point, error) {
	if s == nil {
		return nil, group.ErrNilParam
	}

	s1, ok := s.(*scalar.KyberScalar)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	po2 := KyberPoint{Point: po.Point.Clone()}
	_ = po2.Point.Mul(s1.Scalar, po2.Point)

	return &po2, nil
}

// Pick returns a new random or pseudo-random Point.
func (po *KyberPoint) Pick(rand cipher.Stream) (crypto.Point, error) {
	if rand == nil {
		return nil, group.ErrNilParam
	}

	po2 := KyberPoint{Point: po.Point.Clone()}
	po2.Point.Pick(rand)

	return &po2, nil
}
