package kv2

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	kgroup "gopkg.in/dedis/kyber.v2/group/edwards25519"
)

type suiteEd25519 struct {
	suite *kgroup.SuiteEd25519
}

// NewBlakeSHA256Ed25519 returns a wrapper over a cipher suite based on package
// gopkg.in/dedis/kyber.v2/xof/blake2xb, SHA-256, and the Ed25519 curve.
func NewBlakeSHA256Ed25519() *suiteEd25519 {
	suite := suiteEd25519{}
	suite.suite = kgroup.NewBlakeSHA256Ed25519()

	return &suite
}

// RandomStream returns a cipher.Stream that returns a key stream
// from crypto/rand.
func (s *suiteEd25519) RandomStream() cipher.Stream {
	return s.suite.RandomStream()
}

// CreatePoint creates a new point
func (s *suiteEd25519) CreatePoint() crypto.Point {
	po := kyberPoint{}
	po.Point = s.suite.Curve.Point()
	po.Point.Base()

	return &po
}

// String returns the string for the group
func (s *suiteEd25519) String() string {
	return s.suite.Curve.String()
}

// ScalarLen returns the maximum length of scalars in bytes
func (s *suiteEd25519) ScalarLen() int {
	return s.suite.Curve.ScalarLen()
}

// CreateScalar creates a new Scalar
func (s *suiteEd25519) CreateScalar() crypto.Scalar {
	sc := kyberScalar{}
	sc.Scalar = s.suite.Curve.Scalar()

	return &sc
}

// PointLen returns the max length of point in nb of bytes
func (s *suiteEd25519) PointLen() int {
	return s.suite.Curve.PointLen()
}

// CreateKey returns a formatted Ed25519 key (avoiding subgroup attack by requiring
// it to be a multiple of 8). NewKey implements the crypto.KeyGenerator interface.
func (s *suiteEd25519) CreateKey(stream cipher.Stream) crypto.Scalar {
	sc := kyberScalar{}
	s1 := s.suite.Curve.NewKey(stream)
	sc.Scalar = s1

	return &sc
}

// GetUnderlyingSuite returns the underlying suite
func (s *suiteEd25519) GetUnderlyingSuite() interface{} {
	return s.suite
}
