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

// KeyGenMock mocks a key generation implementation
type KeyGenMock struct {
	GeneratePairMock            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayMock func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayMock  func(b []byte) (crypto.PublicKey, error)
	CheckPublicKeyValidMock     func(b []byte) error
	SuiteMock                   func() crypto.Suite
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

// IsInterfaceNil returns true if there is no value under the interface
func (privKey *PrivateKeyStub) IsInterfaceNil() bool {
	return privKey == nil
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

// IsInterfaceNil returns true if there is no value under the interface
func (pubKey *PublicKeyStub) IsInterfaceNil() bool {
	return pubKey == nil
}

// GeneratePair generates a pair of private and public keys
func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return keyGen.GeneratePairMock()
}

// PrivateKeyFromByteArray generates the private key from it's byte array representation
func (keyGen *KeyGenMock) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	return keyGen.PrivateKeyFromByteArrayMock(b)
}

// PublicKeyFromByteArray generates a public key from it's byte array representation
func (keyGen *KeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return keyGen.PublicKeyFromByteArrayMock(b)
}

// CheckPublicKeyValid verifies the validity of the public key
func (keyGen *KeyGenMock) CheckPublicKeyValid(b []byte) error {
	return keyGen.CheckPublicKeyValidMock(b)
}

// Suite -
func (keyGen *KeyGenMock) Suite() crypto.Suite {
	return keyGen.SuiteMock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (keyGen *KeyGenMock) IsInterfaceNil() bool {
	return keyGen == nil
}
