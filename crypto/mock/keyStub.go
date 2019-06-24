package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// PrivateKeyStub provides stubs for a PrivateKey implementation
type PrivateKeyStub struct {
	ToByteArrayStub    func() ([]byte, error)
	GeneratePublicStub func() crypto.PublicKey
	ScalarStub         func() crypto.Scalar
	SuiteStub          func() crypto.Suite
}

// PublicKeyStub provides stubs for a PublicKey implementation
type PublicKeyStub struct {
	ToByteArrayStub func() ([]byte, error)
	PointStub       func() crypto.Point
	SuiteStub       func() crypto.Suite
}

// ToByteArray returns the byte array representation of the private key
func (privKey *PrivateKeyStub) ToByteArray() ([]byte, error) {
	return privKey.ToByteArrayStub()
}

// GeneratePublic builds a public key for the current private key
func (privKey *PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	return privKey.GeneratePublicStub()
}

// Suite returns the Suite (curve data) used for this private key
func (privKey *PrivateKeyStub) Suite() crypto.Suite {
	return privKey.SuiteStub()
}

// Scalar returns the Scalar corresponding to this Private Key
func (privKey *PrivateKeyStub) Scalar() crypto.Scalar {
	return privKey.ScalarStub()
}

// ToByteArray returns the byte array representation of the public key
func (pubKey *PublicKeyStub) ToByteArray() ([]byte, error) {
	return pubKey.ToByteArrayStub()
}

// Suite returns the Suite (curve data) used for this private key
func (pubKey *PublicKeyStub) Suite() crypto.Suite {
	return pubKey.SuiteStub()
}

// Point returns the Point corresponding to this Public Key
func (pubKey *PublicKeyStub) Point() crypto.Point {
	return pubKey.PointStub()
}
