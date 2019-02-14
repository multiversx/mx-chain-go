package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

// PrivateKeyMock mocks a private key implementation
type PrivateKeyMock struct {
	SignMock           func(message []byte, signer crypto.SingleSigner) ([]byte, error)
	GeneratePublicMock func() crypto.PublicKey
	ToByteArrayMock    func() ([]byte, error)
	SuiteMock          func() crypto.Suite
	ScalarMock         func() crypto.Scalar
}

// PublicKeyMock mocks a public key implementation
type PublicKeyMock struct {
	ToByteArrayMock func() ([]byte, error)
	VerifyMock      func(data []byte, signature []byte, signer crypto.SingleSigner) error
	SuiteMock       func() crypto.Suite
	PointMock       func() crypto.Point
}

// KeyGenMock mocks a key generation implementation
type KeyGenMock struct {
	GeneratePairMock            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayMock func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayMock  func(b []byte) (crypto.PublicKey, error)
	SuiteMock                   func() crypto.Suite
}

// Sign mocks signing with a private key
func (privKey *PrivateKeyMock) Sign(message []byte, signer crypto.SingleSigner) ([]byte, error) {
	return []byte("signed" + string(message)), nil
}

// GeneratePublic mocks generating a public key from the private key
func (privKey *PrivateKeyMock) GeneratePublic() crypto.PublicKey {
	return privKey.GeneratePublicMock()
}

// ToByteArray mocks converting the private key to a byte array
func (privKey *PrivateKeyMock) ToByteArray() ([]byte, error) {
	return []byte("privateKeyMock"), nil
}

func (privKey *PrivateKeyMock) Suite() crypto.Suite {
	return privKey.SuiteMock()
}

func (privKey *PrivateKeyMock) Scalar() crypto.Scalar {
	return privKey.ScalarMock()
}

// ToByteArray mocks converting a public key to a byte array
func (pubKey *PublicKeyMock) ToByteArray() ([]byte, error) {
	return []byte("publicKeyMock"), nil
}

// Verify mocks verifying a signature with a public key
func (pubKey *PublicKeyMock) Verify(data []byte, signature []byte, signer crypto.SingleSigner) error {
	return pubKey.VerifyMock(data, signature, signer)
}

func (pubKey *PublicKeyMock) Suite() crypto.Suite {
	return pubKey.SuiteMock()
}

func (pubKey *PublicKeyMock) Point() crypto.Point {
	return pubKey.PointMock()
}

// GeneratePair generates a pair of private and public keys
func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return keyGen.GeneratePairMock()
}

// PrivateKeyFromByteArray generates the private key from it's byte array representation
func (keyGen *KeyGenMock) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	return keyGen.PrivateKeyFromByteArrayMock(b)
}

// PublicKeyFromByteArrayMock generate a public key from it's byte array representation
func (keyGen *KeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return keyGen.PublicKeyFromByteArrayMock(b)
}

func (keyGen *KeyGenMock) Suite() crypto.Suite {
	return keyGen.SuiteMock()
}
