package kyber

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"go.dedis.ch/kyber/v3"
)

// kyberScalar is a wrapper over kyber scalar
type kyberScalar struct {
	kyber.Scalar
}

// Equal tests if receiver is equal with the scalar s given as parameter.
// Both scalars need to be derived from the same Group
func (ks *kyberScalar) Equal(s crypto.Scalar) (bool, error) {
	if s == nil || s.IsInterfaceNil() {
		return false, crypto.ErrNilParam
	}

	s2, ok := s.(*kyberScalar)

	if !ok {
		return false, crypto.ErrInvalidParam
	}

	areEqual := ks.Scalar.Equal(s2.Scalar)

	return areEqual, nil
}

// Set sets the receiver to Scalar s given as parameter
func (ks *kyberScalar) Set(s crypto.Scalar) error {
	if s == nil || s.IsInterfaceNil() {
		return crypto.ErrNilParam
	}

	s2, ok := s.(*kyberScalar)

	if !ok {
		return crypto.ErrInvalidParam
	}

	_ = ks.Scalar.Set(s2.Scalar)

	return nil
}

// Clone creates a new Scalar with same value as receiver
func (ks *kyberScalar) Clone() crypto.Scalar {
	s2 := ks.Scalar.Clone()
	s := kyberScalar{Scalar: s2}

	return &s
}

// SetInt64 sets the receiver to a small integer value v given as parameter
func (ks *kyberScalar) SetInt64(v int64) {
	_ = ks.Scalar.SetInt64(v)
}

// Zero returns the the additive identity (0)
func (ks *kyberScalar) Zero() crypto.Scalar {
	s1 := ks.Scalar.Clone()
	s := kyberScalar{Scalar: s1.Zero()}

	return &s
}

// Add returns the modular sum of receiver with scalar s given as parameter
func (ks *kyberScalar) Add(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*kyberScalar)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := kyberScalar{Scalar: ks.Scalar.Clone()}
	_ = s1.Scalar.Add(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Sub returns the modular difference between receiver and scalar s given as parameter
func (ks *kyberScalar) Sub(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*kyberScalar)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := kyberScalar{Scalar: ks.Scalar.Clone()}
	_ = s1.Scalar.Sub(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Neg returns the modular negation of receiver
func (ks *kyberScalar) Neg() crypto.Scalar {
	s1 := ks.Scalar.Clone()
	s := kyberScalar{Scalar: s1}
	_ = s.Scalar.Neg(s1)

	return &s
}

// One sets the receiver to the multiplicative identity (1)
func (ks *kyberScalar) One() crypto.Scalar {
	s1 := ks.Scalar.Clone()
	s := kyberScalar{Scalar: s1}
	_ = s.Scalar.One()

	return &s
}

// Mul returns the modular product of receiver with scalar s given as parameter
func (ks *kyberScalar) Mul(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*kyberScalar)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := kyberScalar{Scalar: ks.Scalar.Clone()}
	_ = s1.Scalar.Mul(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Div returns the modular division between receiver and scalar s given as parameter
func (ks *kyberScalar) Div(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*kyberScalar)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := kyberScalar{Scalar: ks.Scalar.Clone()}
	_ = s1.Scalar.Div(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Inv returns the modular inverse of scalar s given as parameter
func (ks *kyberScalar) Inv(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil || s.IsInterfaceNil() {
		return nil, crypto.ErrNilParam
	}

	s2, ok := s.(*kyberScalar)

	if !ok {
		return nil, crypto.ErrInvalidParam
	}

	s1 := kyberScalar{Scalar: ks.Scalar.Clone()}
	_ = s1.Scalar.Inv(s2.Scalar)

	return &s1, nil
}

// Pick returns a fresh random or pseudo-random scalar
func (ks *kyberScalar) Pick(rand cipher.Stream) (crypto.Scalar, error) {
	if rand == nil {
		return nil, crypto.ErrNilParam
	}

	s1 := kyberScalar{Scalar: ks.Scalar.Clone()}
	_ = s1.Scalar.Pick(rand)

	return &s1, nil
}

// SetBytes sets the scalar from a byte-slice,
// reducing if necessary to the appropriate modulus.
func (ks *kyberScalar) SetBytes(s []byte) (crypto.Scalar, error) {
	if s == nil {
		return nil, crypto.ErrNilParam
	}

	s1 := kyberScalar{Scalar: ks.Scalar.Clone()}
	_ = s1.Scalar.SetBytes(s)

	return &s1, nil
}

// GetUnderlyingObj returns the object the implementation wraps
func (ks *kyberScalar) GetUnderlyingObj() interface{} {
	return ks.Scalar
}

// MarshalBinary encodes the receiver into a binary form and returns the result.
func (ks *kyberScalar) MarshalBinary() ([]byte, error) {
	return ks.Scalar.MarshalBinary()
}

// UnmarshalBinary decodes a scalar from its byte array representation and sets the receiver to this value
func (ks *kyberScalar) UnmarshalBinary(s []byte) error {
	return ks.Scalar.UnmarshalBinary(s)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ks *kyberScalar) IsInterfaceNil() bool {
	return ks == nil
}
