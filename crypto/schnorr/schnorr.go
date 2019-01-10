package schnorr

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/pkg/errors"
	"gopkg.in/dedis/kyber.v2"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
	"gopkg.in/dedis/kyber.v2/sign/schnorr"
	"gopkg.in/dedis/kyber.v2/util/key"
)

type keyGenerator struct {
	suite key.Suite
}

// NewKeyGenerator returns a new key generator with the Ed25519 curve suite
func NewKeyGenerator() *keyGenerator {
	return &keyGenerator{suite: edwards25519.NewBlakeSHA256Ed25519()}
}

// privateKey holds the private key and the chosen curve
type privateKey struct {
	suite key.Suite
	sk    kyber.Scalar
}

// publicKey holds the public key and the chosen curve
type publicKey struct {
	suite key.Suite
	pk    kyber.Point
}

// GeneratePair will generate a bundle of private and public key
func (kg *keyGenerator) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	schnorrKeyPair := key.NewKeyPair(kg.suite)
	return &privateKey{
		suite: kg.suite,
		sk:    schnorrKeyPair.Private,
	},
		&publicKey{
			suite: kg.suite,
			pk:    schnorrKeyPair.Public,
		}
}

// PrivateKeyFromByteArray generates a private key given a byte array
func (kg *keyGenerator) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	if b == nil {
		return nil, errors.New("cannot create private key from nil byte array")
	}
	scalar := kg.suite.Scalar()
	err := scalar.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	return &privateKey{
		suite: kg.suite,
		sk:    scalar,
	}, nil
}

// PublicKeyFromByteArray unmarshalls a byte array into a public key Point
func (kg *keyGenerator) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	if b == nil {
		return nil, errors.New("cannot create public key from nil byte array")
	}
	point := kg.suite.Point()
	err := point.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	return &publicKey{
		suite: kg.suite,
		pk:    point,
	}, nil
}

// Sign creates a signature of the message using the current private key
func (spk *privateKey) Sign(message []byte) ([]byte, error) {
	return schnorr.Sign(spk.suite, spk.sk, message)
}

// ToByteArray returns the byte array representation of the private key
func (spk *privateKey) ToByteArray() ([]byte, error) {
	return spk.sk.MarshalBinary()
}

// GeneratePublic builds a public key for the current private key
func (spk *privateKey) GeneratePublic() crypto.PublicKey {
	point := spk.suite.Point()
	return &publicKey{
		suite: spk.suite,
		pk:    point.Mul(spk.sk, nil),
	}
}

// Verify checks a signature over a message
func (pk *publicKey) Verify(data []byte, signature []byte) error {
	err := schnorr.Verify(pk.suite, pk.pk, data, signature)
	if err != nil {
		return err
	}
	return nil
}

// ToByteArray returns the byte array representation of the public key
func (pk *publicKey) ToByteArray() ([]byte, error) {
	return pk.pk.MarshalBinary()
}
