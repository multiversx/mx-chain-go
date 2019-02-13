package signing

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type keyGenerator struct {
	suite crypto.Suite
}

// Pair represents a public/private keypair
type Pair struct {
	Public  crypto.Point
	Private crypto.Scalar
}

// NewKeyGenerator returns a new key generator with the given curve suite
func NewKeyGenerator(suite crypto.Suite) *keyGenerator {
	return &keyGenerator{suite: suite}
}

// GeneratePair will generate a bundle of private and public key
func (kg *keyGenerator) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	keyPair, err := newKeyPair(kg.suite)

	if err != nil {
		panic("unable to generate private/public keys")
	}

	return &privateKey{
			suite: kg.suite,
			sk:    keyPair.Private,
		}, &publicKey{
			suite: kg.suite,
			pk:    keyPair.Public,
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

func newKeyPair(suite crypto.Suite) (*Pair, error) {
	if suite == nil {
		return nil, crypto.ErrNilSuite
	}

	p := new(Pair)
	random := suite.RandomStream()

	if g, ok := suite.(crypto.Generator); ok {
		p.Private = g.CreateKey(random)
	} else {
		privateKey, _ := suite.CreateScalar().Pick(random)
		p.Private = privateKey
	}

	pubKey, _ := suite.CreatePoint().Mul(p.Private)
	p.Public = pubKey

	return p, nil
}
