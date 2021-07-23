package signing

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

var log = logger.GetOrCreate("crypto/signing")

var _ crypto.KeyGenerator = (*keyGenerator)(nil)
var _ crypto.PublicKey = (*publicKey)(nil)
var _ crypto.PrivateKey = (*privateKey)(nil)

// privateKey holds the private key and the chosen curve
type privateKey struct {
	suite crypto.Suite
	sk    crypto.Scalar
}

// publicKey holds the public key and the chosen curve
type publicKey struct {
	suite crypto.Suite
	pk    crypto.Point
}

// keyGenerator generates private and public keys
type keyGenerator struct {
	suite crypto.Suite
}

// NewKeyGenerator returns a new key generator with the given curve suite
func NewKeyGenerator(suite crypto.Suite) *keyGenerator {
	return &keyGenerator{suite: suite}
}

// GeneratePair will generate a bundle of private and public key
func (kg *keyGenerator) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	private, public, err := newKeyPair(kg.suite)

	if err != nil {
		panic("unable to generate private/public keys")
	}

	return &privateKey{
			suite: kg.suite,
			sk:    private,
		}, &publicKey{
			suite: kg.suite,
			pk:    public,
		}
}

// PrivateKeyFromByteArray generates a private key given a byte array
func (kg *keyGenerator) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	if len(b) == 0 {
		return nil, crypto.ErrInvalidParam
	}
	sc := kg.suite.CreateScalar()
	err := sc.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}

	return &privateKey{
		suite: kg.suite,
		sk:    sc,
	}, nil
}

// PublicKeyFromByteArray unmarshalls a byte array into a public key Point
func (kg *keyGenerator) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	if len(b) != kg.suite.PointLen() {
		return nil, crypto.ErrInvalidParam
	}

	point := kg.suite.CreatePoint()
	err := point.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}

	return &publicKey{
		suite: kg.suite,
		pk:    point,
	}, nil
}

// CheckPublicKeyValid verifies the validity of the public key
func (kg *keyGenerator) CheckPublicKeyValid(b []byte) error {
	return kg.suite.CheckPointValid(b)
}

// Suite returns the Suite (curve data) used for this key generator
func (kg *keyGenerator) Suite() crypto.Suite {
	return kg.suite
}

// IsInterfaceNil returns true if there is no value under the interface
func (kg *keyGenerator) IsInterfaceNil() bool {
	return kg == nil
}

func newKeyPair(suite crypto.Suite) (private crypto.Scalar, public crypto.Point, err error) {
	if check.IfNil(suite) {
		return nil, nil, crypto.ErrNilSuite
	}

	private, public = suite.CreateKeyPair()

	return private, public, nil
}

// ToByteArray returns the byte array representation of the private key
func (spk *privateKey) ToByteArray() ([]byte, error) {
	return spk.sk.MarshalBinary()
}

// GeneratePublic builds a public key for the current private key
func (spk *privateKey) GeneratePublic() crypto.PublicKey {
	pubKeyPoint, err := spk.suite.CreatePointForScalar(spk.sk)
	if err != nil {
		log.Warn("problem generating public key",
			"message", err.Error())
	}

	return &publicKey{
		suite: spk.suite,
		pk:    pubKeyPoint,
	}
}

// Suite returns the Suite (curve data) used for this private key
func (spk *privateKey) Suite() crypto.Suite {
	return spk.suite
}

// Scalar returns the Scalar corresponding to this Private Key
func (spk *privateKey) Scalar() crypto.Scalar {
	return spk.sk
}

// IsInterfaceNil returns true if there is no value under the interface
func (spk *privateKey) IsInterfaceNil() bool {
	return spk == nil
}

// ToByteArray returns the byte array representation of the public key
func (pk *publicKey) ToByteArray() ([]byte, error) {
	return pk.pk.MarshalBinary()
}

// Suite returns the Suite (curve data) used for this private key
func (pk *publicKey) Suite() crypto.Suite {
	return pk.suite
}

// Point returns the Point corresponding to this Public Key
func (pk *publicKey) Point() crypto.Point {
	return pk.pk
}

// IsInterfaceNil returns true if there is no value under the interface
func (pk *publicKey) IsInterfaceNil() bool {
	return pk == nil
}
