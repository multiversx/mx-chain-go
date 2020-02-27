package mcl

import (
	"crypto/cipher"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/bls-go-binary/bls"
)

// PointG2 -
type PointG2 struct {
	*bls.G2
}

// NewPointG2 creates a new point on G2 initialized with base point
func NewPointG2() *PointG2 {
	point := &PointG2{
		G2: &bls.G2{},
	}

	basePointG2Str = BaseG2()
	err := point.G2.SetString(basePointG2Str, 10)
	if err != nil {
		fmt.Println(err.Error())
	}

	return point
}

// Equal tests if receiver is equal with the Point p given as parameter.
// Both Points need to be derived from the same Group
func (po *PointG2) Equal(p crypto.Point) (bool, error) {
	if p == nil {
		return false, crypto.ErrNilParam
	}

	po2, ok := p.(*PointG2)
	if !ok {
		return false, crypto.ErrInvalidParam
	}

	return po.G2.IsEqual(po2.G2), nil
}

// Clone returns a clone of the receiver.
func (po *PointG2) Clone() crypto.Point {
	po2 := PointG2{
		G2: &bls.G2{},
	}

	strPo := po.G2.GetString(16)
	_ = po2.G2.SetString(strPo, 16)

	return &po2
}

// Null returns the neutral identity element.
func (po *PointG2) Null() crypto.Point {
	p := &PointG2{
		G2: &bls.G2{},
	}

	p.G2.Clear()

	return p
}

// Set sets the receiver equal to another Point p.
func (po *PointG2) Set(p crypto.Point) error {
	if p == nil {
		return crypto.ErrNilParam
	}

	po1, ok := p.(*PointG2)
	if !ok {
		return crypto.ErrInvalidParam
	}

	strPo := po1.G2.GetString(16)
	_ = po.G2.SetString(strPo, 16)

	return nil
}

// Add returns the result of adding receiver with Point p given as parameter,
// so that their scalars add homomorphically
func (po *PointG2) Add(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*PointG2)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := PointG2{
		G2: &bls.G2{},
	}

	bls.G2Add(po2.G2, po.G2, po1.G2)

	return &po2, nil
}

// Sub returns the result of subtracting from receiver the Point p given as parameter,
// so that their scalars subtract homomorphically
func (po *PointG2) Sub(p crypto.Point) (crypto.Point, error) {
	if p == nil {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*PointG2)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := PointG2{
		G2: &bls.G2{},
	}

	bls.G2Sub(po2.G2, po.G2, po1.G2)

	return &po2, nil
}

// Neg returns the negation of receiver
func (po *PointG2) Neg() crypto.Point {
	po2 := PointG2{
		G2: &bls.G2{},
	}

	bls.G2Neg(po2.G2, po.G2)

	return &po2
}

// Mul returns the result of multiplying receiver by the scalarInt s.
func (po *PointG2) Mul(s crypto.Scalar) (crypto.Point, error) {
	if s == nil {
		return nil, crypto.ErrNilParam
	}

	po2 := PointG2{
		G2: &bls.G2{},
	}

	s1, ok := s.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	bls.G2Mul(po2.G2, po.G2, s1.Scalar)

	return &po2, nil
}

// Pick returns a new random or pseudo-random Point.
func (po *PointG2) Pick(_ cipher.Stream) (crypto.Point, error) {
	scalar := &bls.Fr{}
	scalar.SetByCSPRNG()

	po2 := PointG2{
		G2: &bls.G2{},
	}

	bls.G2Mul(po2.G2, po.G2, scalar)

	return &po2, nil
}

// GetUnderlyingObj returns the object the implementation wraps
func (po *PointG2) GetUnderlyingObj() interface{} {
	return po.G2
}

// MarshalBinary converts the point into its byte array representation
func (po *PointG2) MarshalBinary() ([]byte, error) {
	return po.G2.Serialize(), nil
}

// UnmarshalBinary reconstructs a point from its byte array representation
func (po *PointG2) UnmarshalBinary(point []byte) error {
	return po.G2.Deserialize(point)
}

// IsInterfaceNil returns true if there is no value under the interface
func (po *PointG2) IsInterfaceNil() bool {
	return po == nil
}
