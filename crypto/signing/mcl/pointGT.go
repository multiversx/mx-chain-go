package mcl

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

type PointGT struct {
	*bls.GT
}

// creates a new point on GT initialized with identity
func NewPointGT() *PointGT {
	point := &PointGT{
		GT: &bls.GT{},
	}

	return point
}

// Equal tests if receiver is equal with the Point p given as parameter.
// Both Points need to be derived from the same Group
func (po *PointGT) Equal(p crypto.Point) (bool, error) {
	if p == nil {
		return false, crypto.ErrNilParam
	}

	po2, ok := p.(*PointGT)
	if !ok {
		return false, crypto.ErrInvalidParam
	}

	return po.GT.IsEqual(po2.GT), nil
}

// Clone returns a clone of the receiver.
func (po *PointGT) Clone() crypto.Point {
	po2 := PointGT{
		GT: &bls.GT{},
	}
	_ = po2.Deserialize(po.GT.Serialize())

	return &po2
}

// Null returns the neutral identity element.
func (po *PointGT) Null() crypto.Point {
	return NewPointGT()
}

// Set sets the receiver equal to another Point p.
func (po *PointGT) Set(p crypto.Point) error {
	if p == nil {
		return crypto.ErrNilParam
	}

	po1, ok := p.(*PointGT)
	if !ok {
		return crypto.ErrInvalidParam
	}

	point := po1.Clone().(*PointGT)
	po.GT = point.GT

	return nil
}

// Add returns the result of adding receiver with Point p given as parameter,
// so that their scalars add homomorphically
func (po *PointGT) Add(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*PointGT)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := PointGT{
		GT: &bls.GT{},
	}

	bls.GTAdd(po2.GT, po.GT, po1.GT)

	return &po2, nil
}

// Sub returns the result of subtracting from receiver the Point p given as parameter,
// so that their scalars subtract homomorphically
func (po *PointGT) Sub(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*PointGT)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := PointGT{
		GT: &bls.GT{},
	}

	bls.GTSub(po2.GT, po.GT, po1.GT)

	return &po2, nil
}

// Neg returns the negation of receiver
func (po *PointGT) Neg() crypto.Point {
	po2 := PointGT{
		GT: &bls.GT{},
	}

	bls.GTNeg(po2.GT, po.GT)

	return &po2
}

// Mul returns the result of multiplying receiver by the scalarInt s.
func (po *PointGT) Mul(s crypto.Scalar) (crypto.Point, error) {
	if s == nil {
		return nil, crypto.ErrNilParam
	}

	s1, ok := s.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := PointGT{
		GT: &bls.GT{},
	}

	bls.GTPow(po2.GT, po.GT, s1.Scalar)

	return &po2, nil
}

// Pick returns a new random or pseudo-random Point.
func (po *PointGT) Pick(rand cipher.Stream) (crypto.Point, error) {
	var p1, p2 crypto.Point
	var err error

	p1, err = NewPointG1().Pick(rand)
	if err != nil {
		return nil, err
	}

	p2, err = NewPointG2().Pick(rand)
	if err != nil {
		return nil, err
	}

	poG1 := p1.(*PointG1)
	poG2 := p2.(*PointG2)

	po2 := PointGT{
		GT: &bls.GT{},
	}

	bls.Pairing(po2.GT, poG1.G1, poG2.G2)

	return &po2, nil
}

// GetUnderlyingObj returns the object the implementation wraps
func (po *PointGT) GetUnderlyingObj() interface{} {
	return po.GT
}

// MarshalBinary converts the point into its byte array representation
func (po *PointGT) MarshalBinary() ([]byte, error) {
	return po.GT.Serialize(), nil
}

// UnmarshalBinary reconstructs a point from its byte array representation
func (po *PointGT) UnmarshalBinary(point []byte) error {
	return po.GT.Deserialize(point)
}

// IsInterfaceNil returns true if there is no value under the interface
func (po *PointGT) IsInterfaceNil() bool {
	return po == nil
}
