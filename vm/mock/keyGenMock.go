package mock

import "github.com/multiversx/mx-chain-crypto-go"

// KeyGenMock -
type KeyGenMock struct {
	PublicKeyFromByteArrayCalled func(b []byte) (crypto.PublicKey, error)
}

// PublicKeyMock mocks a public key implementation
type PublicKeyMock struct {
}

// ToByteArray mocks converting a public key to a byte array
func (pubKey *PublicKeyMock) ToByteArray() ([]byte, error) {
	return []byte("publicKeyMock"), nil
}

// Suite -
func (pubKey *PublicKeyMock) Suite() crypto.Suite {
	return nil
}

// Point -
func (pubKey *PublicKeyMock) Point() crypto.Point {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pubKey *PublicKeyMock) IsInterfaceNil() bool {
	return pubKey == nil
}

// GeneratePair -
func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return nil, nil
}

// PrivateKeyFromByteArray -
func (keyGen *KeyGenMock) PrivateKeyFromByteArray(_ []byte) (crypto.PrivateKey, error) {
	return nil, nil
}

// PublicKeyFromByteArray -
func (keyGen *KeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	if keyGen.PublicKeyFromByteArrayCalled == nil {
		return nil, nil
	}
	return keyGen.PublicKeyFromByteArrayCalled(b)
}

// CheckPublicKeyValid -
func (keyGen *KeyGenMock) CheckPublicKeyValid(_ []byte) error {
	return nil
}

// Suite -
func (keyGen *KeyGenMock) Suite() crypto.Suite {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (keyGen *KeyGenMock) IsInterfaceNil() bool {
	return keyGen == nil
}
