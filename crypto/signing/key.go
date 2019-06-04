package signing

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

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
	if b == nil {
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
	if b == nil {
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

// Suite returns the Suite (curve data) used for this key generator
func (kg *keyGenerator) Suite() crypto.Suite {
	return kg.suite
}

func newKeyPair(suite crypto.Suite) (private crypto.Scalar, public crypto.Point, err error) {
	if suite == nil {
		return nil, nil, crypto.ErrNilSuite
	}

	random := suite.RandomStream()

	private, public = suite.CreateKeyPair(random)

	return private, public, nil
}

// ToByteArray returns the byte array representation of the private key
func (spk *privateKey) ToByteArray() ([]byte, error) {
	return spk.sk.MarshalBinary()
}

// GeneratePublic builds a public key for the current private key
func (spk *privateKey) GeneratePublic() crypto.PublicKey {
	point := spk.suite.CreatePoint().Base()
	pubKeyPoint, _ := point.Mul(spk.sk)
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
