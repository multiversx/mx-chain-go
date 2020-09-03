package ed25519

import (
	"crypto/cipher"
	"crypto/ed25519"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

var log = logger.GetOrCreate("crypto/signing/ed25519")

var _ crypto.Group = (*suiteEd25519)(nil)
var _ crypto.Random = (*suiteEd25519)(nil)
var _ crypto.Suite = (*suiteEd25519)(nil)

// ED25519 is the string representations of the ed25519 scheme
const ED25519 = "Ed25519"

type suiteEd25519 struct{}

// NewEd25519 returns a wrapper over ed25519
func NewEd25519() *suiteEd25519 {
	return &suiteEd25519{}
}

// CreateKeyPair returns a pair of Ed25519 keys
func (s *suiteEd25519) CreateKeyPair() (crypto.Scalar, crypto.Point) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic("could not create ed25519 key pair: " + err.Error())
	}

	return &ed25519Scalar{privateKey}, &ed25519Point{publicKey}
}

// CreatePoint returns a newly created public key which is a point on ed25519
func (s *suiteEd25519) CreatePoint() crypto.Point {
	_, publicKey := s.CreateKeyPair()
	return publicKey
}

// CreatePointForScalar returns a Point that is the representation of a public key corresponding
//  to the provided Scalar in the ed25519 signature scheme
func (s *suiteEd25519) CreatePointForScalar(scalar crypto.Scalar) (crypto.Point, error) {
	privateKey, ok := scalar.GetUnderlyingObj().(ed25519.PrivateKey)
	if !ok {
		return nil, crypto.ErrInvalidPrivateKey
	}

	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, crypto.ErrGeneratingPubFromPriv
	}

	return &ed25519Point{publicKey}, nil
}

// String returns the string for the group
func (s *suiteEd25519) String() string {
	return ED25519
}

// ScalarLen returns the length of the ed25519 private key - which is the number of bytes of
//  the seed (the actual private key) + the number of bytes of the public key
func (s *suiteEd25519) ScalarLen() int {
	return ed25519.PrivateKeySize
}

// CreateScalar creates a new Scalar which represents the ed25519 private key
func (s *suiteEd25519) CreateScalar() crypto.Scalar {
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic("could not create ed25519 private key: " + err.Error())
	}

	return &ed25519Scalar{privateKey}
}

// PointLen returns the number of bytes of the ed25519 public key
func (s *suiteEd25519) PointLen() int {
	return ed25519.PublicKeySize
}

// GetUnderlyingSuite returns nothing because this is not a wrapper over another suite implementation
func (s *suiteEd25519) GetUnderlyingSuite() interface{} {
	log.Warn("suiteEd25519",
		"message", "calling GetUnderlyingSuite for suiteEd25519 which has no underlying suite")

	return nil
}

// IsPointValid -
func (s *suiteEd25519) CheckPointValid(pointBytes []byte) error {
	if len(pointBytes) != s.PointLen() {
		return crypto.ErrInvalidParam
	}

	point := s.CreatePoint()
	err := point.UnmarshalBinary(pointBytes)
	if err != nil {
		return err
	}

	return nil
}

// RandomStream returns nothing - TODO: Remove this
func (s *suiteEd25519) RandomStream() cipher.Stream {
	log.Debug("suiteEd25519",
		"message", "calling RandomStream for suiteEd25519 - this function should not be used")

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *suiteEd25519) IsInterfaceNil() bool {
	return s == nil
}
