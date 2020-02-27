package mcl

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

// MclScalar -
type MclScalar struct {
	Scalar *bls.Fr
}

// NewMclScalar --
func NewMclScalar() *MclScalar {
	scalar := &MclScalar{Scalar: &bls.Fr{}}
	scalar.Scalar.SetByCSPRNG()
	for scalar.Scalar.IsOne() || scalar.Scalar.IsZero() {
		scalar.Scalar.SetByCSPRNG()
	}

	return scalar
}

// Equal tests if receiver is equal with the scalarInt s given as parameter.
// Both scalars need to be derived from the same Group
func (sc *MclScalar) Equal(s crypto.Scalar) (bool, error) {
	if s == nil || s.IsInterfaceNil() {
		return false, crypto.ErrNilParam
	}

	s2, ok := s.(*MclScalar)
	if !ok {
		return false, crypto.ErrInvalidParam
	}

	areEqual := sc.Scalar.IsEqual(s2.Scalar)

	return areEqual, nil
}

// Set sets the receiver to Scalar s given as parameter
func (sc *MclScalar) Set(s crypto.Scalar) error {
	if s == nil || s.IsInterfaceNil() {
		return crypto.ErrNilParam
	}

	s2, ok := s.(*MclScalar)
	if !ok {
		return crypto.ErrInvalidParam
	}

	return sc.Scalar.Deserialize(s2.Scalar.Serialize())
}

// Clone creates a new Scalar with same value as receiver
func (sc *MclScalar) Clone() crypto.Scalar {
	s := MclScalar{
		Scalar: &bls.Fr{},
	}

	_ = s.Scalar.Deserialize(sc.Scalar.Serialize())

	return &s
}

// SetInt64 sets the receiver to a small integer value v given as parameter
func (sc *MclScalar) SetInt64(v int64) {
	sc.Scalar.SetInt64(v)
}

// Zero returns the the additive identity (0)
func (sc *MclScalar) Zero() crypto.Scalar {
	s := MclScalar{
		Scalar: &bls.Fr{},
	}
	s.Scalar.Clear()

	return &s
}

// Add returns the modular sum of receiver with scalarInt s given as parameter
func (sc *MclScalar) Add(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := MclScalar{
		Scalar: &bls.Fr{},
	}

	bls.FrAdd(s1.Scalar, sc.Scalar, s2.Scalar)

	return &s1, nil
}

// Sub returns the modular difference between receiver and scalarInt s given as parameter
func (sc *MclScalar) Sub(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := MclScalar{
		Scalar: &bls.Fr{},
	}

	bls.FrSub(s1.Scalar, sc.Scalar, s2.Scalar)

	return &s1, nil
}

// Neg returns the modular negation of receiver
func (sc *MclScalar) Neg() crypto.Scalar {
	s := MclScalar{
		Scalar: &bls.Fr{},
	}

	bls.FrNeg(s.Scalar, sc.Scalar)

	return &s
}

// One sets the receiver to the multiplicative identity (1)
func (sc *MclScalar) One() crypto.Scalar {
	s := MclScalar{
		Scalar: &bls.Fr{},
	}

	s.Scalar.SetInt64(1)

	return &s
}

// Mul returns the modular product of receiver with scalarInt s given as parameter
func (sc *MclScalar) Mul(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := MclScalar{
		Scalar: &bls.Fr{},
	}

	bls.FrMul(s1.Scalar, sc.Scalar, s2.Scalar)

	return &s1, nil
}

// Div returns the modular division between receiver and scalarInt s given as parameter
func (sc *MclScalar) Div(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := MclScalar{
		Scalar: &bls.Fr{},
	}

	bls.FrDiv(s1.Scalar, sc.Scalar, s2.Scalar)

	return &s1, nil
}

// Inv returns the modular inverse of scalarInt s given as parameter
func (sc *MclScalar) Inv(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := MclScalar{
		Scalar: &bls.Fr{},
	}

	bls.FrInv(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Pick returns a fresh random or pseudo-random scalarInt
func (sc *MclScalar) Pick(_ cipher.Stream) (crypto.Scalar, error) {
	s1 := MclScalar{
		Scalar: &bls.Fr{},
	}

	s1.Scalar.SetByCSPRNG()

	return &s1, nil
}

// SetBytes sets the scalarInt from a byte-slice,
// reducing if necessary to the appropriate modulus.
func (sc *MclScalar) SetBytes(s []byte) (crypto.Scalar, error) {
	if s == nil {
		return nil, crypto.ErrNilParam
	}

	s1 := sc.Clone()
	s1Mcl, ok := s1.(*MclScalar)
	if !ok {
		return nil, crypto.ErrInvalidScalar
	}

	err := s1Mcl.Scalar.Deserialize(s)
	if err != nil {
		return nil, err
	}

	return s1, nil
}

// GetUnderlyingObj returns the object the implementation wraps
func (sc *MclScalar) GetUnderlyingObj() interface{} {
	return sc.Scalar
}

// MarshalBinary encodes the receiver into a binary form and returns the result.
func (sc *MclScalar) MarshalBinary() ([]byte, error) {
	return sc.Scalar.Serialize(), nil
}

// UnmarshalBinary decodes a scalarInt from its byte array representation and sets the receiver to this value
func (sc *MclScalar) UnmarshalBinary(s []byte) error {
	return sc.Scalar.Deserialize(s)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *MclScalar) IsInterfaceNil() bool {
	return sc == nil
}
