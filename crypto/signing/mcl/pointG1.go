package mcl

import (
	"runtime"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

// PointG1 -
type PointG1 struct {
	*bls.G1
}

// NewPointG1 creates a new point on G1 initialized with base point
func NewPointG1() *PointG1 {
	point := &PointG1{
		G1: &bls.G1{},
	}

	bpG1Str := baseG1()
	err := point.G1.SetString(bpG1Str, 10)
	if err != nil {
		panic(err.Error())
	}

	return point
}

// Equal tests if receiver is equal with the Point p given as parameter.
// Both Points need to be derived from the same Group
func (po *PointG1) Equal(p crypto.Point) (bool, error) {
	if p == nil {
		return false, crypto.ErrNilParam
	}

	po2, ok := p.(*PointG1)
	if !ok {
		return false, crypto.ErrInvalidParam
	}

	return po.G1.IsEqual(po2.G1), nil
}

// Clone returns a clone of the receiver.
func (po *PointG1) Clone() crypto.Point {
	po2 := PointG1{
		G1: &bls.G1{},
	}

	strPo := po.G1.GetString(16)
	err := po2.G1.SetString(strPo, 16)
	if err != nil {
		log.Error("PointG1 Clone", "error", err.Error())
	}

	return &po2
}

// Null returns the neutral identity element.
func (po *PointG1) Null() crypto.Point {
	p := &PointG1{
		G1: &bls.G1{},
	}

	p.G1.Clear()

	return p
}

// Set sets the receiver equal to another Point p.
func (po *PointG1) Set(p crypto.Point) error {
	if check.IfNil(p) {
		return crypto.ErrNilParam
	}

	po1, ok := p.(*PointG1)
	if !ok {
		return crypto.ErrInvalidParam
	}

	strPo := po1.G1.GetString(16)
	return po.G1.SetString(strPo, 16)
}

// Add returns the result of adding receiver with Point p given as parameter,
// so that their scalars add homomorphically
func (po *PointG1) Add(p crypto.Point) (crypto.Point, error) {
	if check.IfNil(p) {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*PointG1)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := PointG1{
		G1: &bls.G1{},
	}

	bls.G1Add(po2.G1, po.G1, po1.G1)
	runtime.KeepAlive(po1)

	return &po2, nil
}

// Sub returns the result of subtracting from receiver the Point p given as parameter,
// so that their scalars subtract homomorphically
func (po *PointG1) Sub(p crypto.Point) (crypto.Point, error) {
	if check.IfNil(p) {
		return nil, crypto.ErrNilParam
	}

	po1, ok := p.(*PointG1)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	po2 := PointG1{
		G1: &bls.G1{},
	}

	bls.G1Sub(po2.G1, po.G1, po1.G1)
	runtime.KeepAlive(po1)

	return &po2, nil
}

// Neg returns the negation of receiver
func (po *PointG1) Neg() crypto.Point {
	po2 := PointG1{
		G1: &bls.G1{},
	}

	bls.G1Neg(po2.G1, po.G1)

	return &po2
}

// Mul returns the result of multiplying receiver by the scalarInt s.
func (po *PointG1) Mul(s crypto.Scalar) (crypto.Point, error) {
	if check.IfNil(s) {
		return nil, crypto.ErrNilParam
	}

	po2 := PointG1{
		G1: &bls.G1{},
	}

	s1, ok := s.(*Scalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	bls.G1MulCT(po2.G1, po.G1, s1.Scalar)
	runtime.KeepAlive(s1)

	return &po2, nil
}

// Pick returns a new random or pseudo-random Point.
func (po *PointG1) Pick() (crypto.Point, error) {
	scalar := &bls.Fr{}
	scalar.SetByCSPRNG()

	po2 := PointG1{
		G1: &bls.G1{},
	}

	bls.G1MulCT(po2.G1, po.G1, scalar)
	runtime.KeepAlive(scalar)

	return &po2, nil
}

// GetUnderlyingObj returns the object the implementation wraps
func (po *PointG1) GetUnderlyingObj() interface{} {
	return po.G1
}

// MarshalBinary converts the point into its byte array representation
func (po *PointG1) MarshalBinary() ([]byte, error) {
	return po.G1.Serialize(), nil
}

// UnmarshalBinary reconstructs a point from its byte array representation
func (po *PointG1) UnmarshalBinary(point []byte) error {
	return po.G1.Deserialize(point)
}

// IsInterfaceNil returns true if there is no value under the interface
func (po *PointG1) IsInterfaceNil() bool {
	return po == nil
}
