package mcl

import (
	"crypto/cipher"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

type InternalSuite struct {
	G1 groupG1
	G2 groupG2
	GT groupGT
}

type suiteBLS12 struct {
	strSuite string
	suite    InternalSuite
}

func init() {
	if err := bls.Init(bls.BLS12_381); err != nil {
		panic(fmt.Sprintf("could not initialize BLS12-381 curve %v", err))
	}
}

// NewSuiteBLS12 returns a wrapper over a BLS12 curve.
func NewSuiteBLS12() *suiteBLS12 {
	return &suiteBLS12{
		strSuite: "BLS12-381 suite",
	}
}

// RandomStream returns a cipher.Stream that returns a key stream
// from crypto/rand.
func (s *suiteBLS12) RandomStream() cipher.Stream {
	panic("not supported")
}

// CreatePoint creates a new point
func (s *suiteBLS12) CreatePoint() crypto.Point {
	return s.suite.G2.CreatePoint()
}

// String returns the string for the group
func (s *suiteBLS12) String() string {
	return s.strSuite
}

// ScalarLen returns the maximum length of scalars in bytes
func (s *suiteBLS12) ScalarLen() int {
	return s.suite.G2.ScalarLen()
}

// CreateScalar creates a new Scalar
func (s *suiteBLS12) CreateScalar() crypto.Scalar {
	return s.suite.G2.CreateScalar()
}

// CreatePointForScalar creates a new point corresponding to the given scalar
func (s *suiteBLS12) CreatePointForScalar(scalar crypto.Scalar) crypto.Point {
	return s.suite.G2.CreatePointForScalar(scalar)
}

// PointLen returns the max length of point in nb of bytes
func (s *suiteBLS12) PointLen() int {
	return s.suite.G2.PointLen()
}

// CreateKeyPair returns a pair of private public BLS keys.
// The private key is a scalarInt, while the public key is a Point on G2 curve
func (s *suiteBLS12) CreateKeyPair(stream cipher.Stream) (crypto.Scalar, crypto.Point) {
	var sc crypto.Scalar
	var p crypto.Point

	sc = s.suite.G2.CreateScalar()
	sc, _ = sc.Pick(stream)
	p = s.suite.G2.CreatePointForScalar(sc)

	return sc, p
}

// GetUnderlyingSuite returns the underlying suite
func (s *suiteBLS12) GetUnderlyingSuite() interface{} {
	return s.suite
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *suiteBLS12) IsInterfaceNil() bool {
	return s == nil
}
