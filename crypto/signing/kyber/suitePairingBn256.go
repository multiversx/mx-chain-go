package kyber

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"go.dedis.ch/kyber/v3/pairing"
)

type suitePairingBn256 struct {
	suite *pairing.SuiteBn256
}

// NewSuitePairingBn256 returns a wrapper over a Bn256 kyber suite.
func NewSuitePairingBn256() *suitePairingBn256 {
	suite := suitePairingBn256{}
	suite.suite = pairing.NewSuiteBn256()

	return &suite
}

// RandomStream returns a cipher.Stream that returns a key stream
// from crypto/rand.
func (s *suitePairingBn256) RandomStream() cipher.Stream {
	return s.suite.RandomStream()
}

// CreatePoint creates a new point
func (s *suitePairingBn256) CreatePoint() crypto.Point {
	po := kyberPoint{}
	po.Point = s.suite.Point()
	// No Base necessary. For marshal /unmarshal kyber expects an empty point

	return &po
}

// String returns the string for the group
func (s *suitePairingBn256) String() string {
	return s.suite.String()
}

// ScalarLen returns the maximum length of scalars in bytes
func (s *suitePairingBn256) ScalarLen() int {
	return s.suite.ScalarLen()
}

// CreateScalar creates a new Scalar
func (s *suitePairingBn256) CreateScalar() crypto.Scalar {
	sc := kyberScalar{}
	sc.Scalar = s.suite.Scalar()

	return &sc
}

// PointLen returns the max length of point in nb of bytes
func (s *suitePairingBn256) PointLen() int {
	return s.suite.PointLen()
}

// CreateKeyPair returns a pair of private public BLS keys.
// The private key is a scalar, while the public key is a Point on G2 curve
func (s *suitePairingBn256) CreateKeyPair(stream cipher.Stream) (crypto.Scalar, crypto.Point) {
	sc := kyberScalar{}
	p := kyberPoint{}

	s1 := s.suite.G2().Scalar().Pick(stream)
	sc.Scalar = s1

	p1 := s.suite.G2().Point().Mul(s1, nil)
	p.Point = p1

	return &sc, &p
}

// GetUnderlyingSuite returns the underlying suite
func (s *suitePairingBn256) GetUnderlyingSuite() interface{} {
	return s.suite
}
