package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// PrivateKeyMock mocks a private key implementation
type PrivateKeyMock struct {
	GeneratePublicMock func() crypto.PublicKey
	ToByteArrayMock    func() ([]byte, error)
	SuiteMock          func() crypto.Suite
	ScalarMock         func() crypto.Scalar
}

// PublicKeyMock mocks a public key implementation
type PublicKeyMock struct {
	ToByteArrayMock func() ([]byte, error)
	SuiteMock       func() crypto.Suite
	PointMock       func() crypto.Point
}

// KeyGenMock mocks a key generation implementation
type KeyGenMock struct {
	GeneratePairMock            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayMock func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayMock  func(b []byte) (crypto.PublicKey, error)
	CheckPublicKeyValidMock     func(b []byte) error
	SuiteMock                   func() crypto.Suite
}

// GeneratePublic mocks generating a public key from the private key
func (privKey *PrivateKeyMock) GeneratePublic() crypto.PublicKey {
	if privKey.GeneratePublicMock != nil {
		return privKey.GeneratePublicMock()
	}

	return &PublicKeyMock{}
}

// ToByteArray mocks converting the private key to a byte array
func (privKey *PrivateKeyMock) ToByteArray() ([]byte, error) {
	return []byte("privateKeyMock"), nil
}

// Suite -
func (privKey *PrivateKeyMock) Suite() crypto.Suite {
	return privKey.SuiteMock()
}

// Scalar -
func (privKey *PrivateKeyMock) Scalar() crypto.Scalar {
	return privKey.ScalarMock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (privKey *PrivateKeyMock) IsInterfaceNil() bool {
	return privKey == nil
}

// ToByteArray mocks converting a public key to a byte array
func (pubKey *PublicKeyMock) ToByteArray() ([]byte, error) {
	return []byte("publicKeyMock"), nil
}

// Suite -
func (pubKey *PublicKeyMock) Suite() crypto.Suite {
	return pubKey.SuiteMock()
}

// Point -
func (pubKey *PublicKeyMock) Point() crypto.Point {
	return pubKey.PointMock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pubKey *PublicKeyMock) IsInterfaceNil() bool {
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
