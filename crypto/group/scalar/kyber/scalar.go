package kyber

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"gopkg.in/dedis/kyber.v2"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/group"
)

// KyberScalar
type KyberScalar struct {
	kyber.Scalar
}

// Equal tests if receiver is equal with the scalar s given as parameter.
// Both scalars need to be derived from the same Group
func (sc *KyberScalar) Equal(s crypto.Scalar) (bool, error) {
	if s == nil {
		return false, group.ErrNilParam
	}

	s2, ok := s.(*KyberScalar)

	if !ok {
		return false, group.ErrInvalidParam
	}

	areEqual := sc.Scalar.Equal(s2.Scalar)

	return areEqual, nil
}

// Set sets the receiver to Scalar s given as parameter
func (sc *KyberScalar) Set(s crypto.Scalar) error {
	if s == nil {
		return group.ErrNilParam
	}

	s2, ok := s.(*KyberScalar)

	if !ok {
		return group.ErrInvalidParam
	}

	_ = sc.Scalar.Set(s2.Scalar)

	return nil
}

// Clone creates a new Scalar with same value as receiver
func (sc *KyberScalar) Clone() crypto.Scalar {
	s2 := sc.Scalar.Clone()
	s := KyberScalar{Scalar: s2}

	return &s
}

// SetInt64 sets the receiver to a small integer value v given as parameter
func (sc *KyberScalar) SetInt64(v int64) {
	_ = sc.Scalar.SetInt64(v)
}

// Zero returns the the additive identity (0)
func (sc *KyberScalar) Zero() crypto.Scalar {
	s1 := sc.Scalar.Clone()
	s := KyberScalar{Scalar: s1.Zero()}

	return &s
}

// Add returns the modular sum of receiver with scalar s given as parameter
func (sc *KyberScalar) Add(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil {
		return nil, group.ErrNilParam
	}

	s2, ok := s.(*KyberScalar)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	s1 := KyberScalar{Scalar: sc.Scalar.Clone()}
	_ = s1.Scalar.Add(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Sub returns the modular difference between receiver and scalar s given as parameter
func (sc *KyberScalar) Sub(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil {
		return nil, group.ErrNilParam
	}

	s2, ok := s.(*KyberScalar)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	s1 := KyberScalar{Scalar: sc.Scalar.Clone()}
	_ = s1.Scalar.Sub(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Neg returns the modular negation of receiver
func (sc *KyberScalar) Neg() crypto.Scalar {
	s1 := sc.Scalar.Clone()
	s := KyberScalar{Scalar: s1}
	_ = s.Scalar.Neg(s1)

	return &s
}

// One sets the receiver to the multiplicative identity (1)
func (sc *KyberScalar) One() crypto.Scalar {
	s1 := sc.Scalar.Clone()
	s := KyberScalar{Scalar: s1}
	_ = s.Scalar.One()

	return &s
}

// Mul returns the modular product of receiver with scalar s given as parameter
func (sc *KyberScalar) Mul(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil {
		return nil, group.ErrNilParam
	}

	s2, ok := s.(*KyberScalar)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	s1 := KyberScalar{Scalar: sc.Scalar.Clone()}
	_ = s1.Scalar.Mul(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Div returns the modular division between receiver and scalar s given as parameter
func (sc *KyberScalar) Div(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil {
		return nil, group.ErrNilParam
	}

	s2, ok := s.(*KyberScalar)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	s1 := KyberScalar{Scalar: sc.Scalar.Clone()}
	_ = s1.Scalar.Div(s1.Scalar, s2.Scalar)

	return &s1, nil
}

// Inv returns the modular inverse of scalar s given as parameter
func (sc *KyberScalar) Inv(s crypto.Scalar) (crypto.Scalar, error) {
	if s == nil {
		return nil, group.ErrNilParam
	}

	s2, ok := s.(*KyberScalar)

	if !ok {
		return nil, group.ErrInvalidParam
	}

	s1 := KyberScalar{Scalar: sc.Scalar.Clone()}
	_ = s1.Scalar.Inv(s2)

	return &s1, nil
}

// Pick returns a fresh random or pseudo-random scalar
func (sc *KyberScalar) Pick(rand cipher.Stream) (crypto.Scalar, error) {
	if rand == nil {
		return nil, group.ErrNilParam
	}

	s1 := KyberScalar{Scalar: sc.Scalar.Clone()}
	_ = s1.Scalar.Pick(rand)

	return &s1, nil
}

// SetBytes sets the scalar from a byte-slice,
// reducing if necessary to the appropriate modulus.
func (sc *KyberScalar) SetBytes(s []byte) (crypto.Scalar, error) {
	if s == nil {
		return nil, group.ErrNilParam
	}

	s1 := KyberScalar{Scalar: sc.Scalar.Clone()}
	_ = s1.Scalar.SetBytes(s)

	return &s1, nil
}

// MarshalBinary encodes the receiver into a binary form and returns the result.
func (sc *KyberScalar) MarshalBinary() ([]byte, error) {
	return sc.Scalar.MarshalBinary()
}
