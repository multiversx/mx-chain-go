package schnorr

import (
	"gopkg.in/dedis/kyber.v2"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
	"gopkg.in/dedis/kyber.v2/sign/schnorr"
	"gopkg.in/dedis/kyber.v2/util/key"
)


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
func GeneratePair() (*privateKey, *publicKey) {
	schnorrSig := edwards25519.NewBlakeSHA256Ed25519()
	schnorrKeyPair := key.NewKeyPair(schnorrSig)
	return &privateKey{
			suite: schnorrSig,
			sk:    schnorrKeyPair.Private,
		},
		&publicKey{
			suite: schnorrSig,
			pk:    schnorrKeyPair.Public,
		}
}

// PrivateKeyFromByteArray generates a private key given a byte array
func PrivateKeyFromByteArray(b []byte) (*privateKey, error) {
	suite :=  edwards25519.NewBlakeSHA256Ed25519()
	scalar := suite.Scalar()
	err := scalar.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	return &privateKey{
		suite: suite,
		sk: scalar,
	}, nil
}

// PublicKeyFromByteArray unmarshalls a byte array into a public key Point
func PublicKeyFromByteArray(b []byte) (*publicKey, error) {
	suite :=  edwards25519.NewBlakeSHA256Ed25519()
	point := suite.Point()
	err := point.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	return &publicKey{
		suite: suite,
		pk: point,
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
func (spk *privateKey) GeneratePublic() *publicKey {
	point := spk.suite.Point()
	return &publicKey{
		suite: spk.suite,
		pk: point.Mul(spk.sk, nil),
	}
}

// Verify checks a signature over a message
func (spk *publicKey) Verify(data []byte, signature []byte) (bool, error) {
	err := schnorr.Verify(spk.suite, spk.pk, data, signature)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ToByteArray returns the byte array representation of the public key
func (spk *publicKey) ToByteArray() ([]byte, error) {
	return spk.pk.MarshalBinary()
}
