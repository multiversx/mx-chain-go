package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

// PrivateKeyMock mocks a private key implementation
type PrivateKeyMock struct {
	SignMock           func(message []byte) ([]byte, error)
	GeneratePublicMock func() crypto.PublicKey
	ToByteArrayMock    func() ([]byte, error)
}

// PublicKeyMock mocks a public key implementation
type PublicKeyMock struct {
	ToByteArrayMock func() ([]byte, error)
	VerifyMock      func(data []byte, signature []byte) (bool, error)
}

// Sign mocks signing with a private key
func (privKey *PrivateKeyMock) Sign(message []byte) ([]byte, error) {
	return privKey.SignMock(message)
}

// GeneratePublic mocks generating a public key from the private key
func (privKey *PrivateKeyMock) GeneratePublic() crypto.PublicKey {
	return privKey.GeneratePublicMock()
}

// ToByteArray mocks converting the private key to a byte array
func (privKey *PrivateKeyMock) ToByteArray() ([]byte, error) {
	return privKey.ToByteArrayMock()
}

// ToByteArray mocks converting a public key to a byte array
func (pubKey *PublicKeyMock) ToByteArray() ([]byte, error) {
	return pubKey.ToByteArrayMock()
}

// Verify mocks verifying a signature with a public key
func (pubKey *PublicKeyMock) Verify(data []byte, signature []byte) (bool, error) {
	return pubKey.VerifyMock(data, signature)
}
