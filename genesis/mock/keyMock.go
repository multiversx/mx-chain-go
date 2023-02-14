package mock

import (
	"github.com/multiversx/mx-chain-crypto-go"
)

// PublicKeyMock -
type PublicKeyMock struct {
}

// PrivateKeyMock -
type PrivateKeyMock struct {
}

// KeyGenMock -
type KeyGenMock struct {
}

// ToByteArray -
func (sspk *PublicKeyMock) ToByteArray() ([]byte, error) {
	return []byte("pubKey"), nil
}

// Suite -
func (sspk *PublicKeyMock) Suite() crypto.Suite {
	return nil
}

// Point -
func (sspk *PublicKeyMock) Point() crypto.Point {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sspk *PublicKeyMock) IsInterfaceNil() bool {
	return sspk == nil
}

// ToByteArray -
func (sk *PrivateKeyMock) ToByteArray() ([]byte, error) {
	return []byte("privKey"), nil
}

// GeneratePublic -
func (sk *PrivateKeyMock) GeneratePublic() crypto.PublicKey {
	return &PublicKeyMock{}
}

// Suite -
func (sk *PrivateKeyMock) Suite() crypto.Suite {
	return nil
}

// Scalar -
func (sk *PrivateKeyMock) Scalar() crypto.Scalar {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sk *PrivateKeyMock) IsInterfaceNil() bool {
	return sk == nil
}

// GeneratePair -
func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return &PrivateKeyMock{}, &PublicKeyMock{}
}

// PrivateKeyFromByteArray -
func (keyGen *KeyGenMock) PrivateKeyFromByteArray(_ []byte) (crypto.PrivateKey, error) {
	return &PrivateKeyMock{}, nil
}

// PublicKeyFromByteArray -
func (keyGen *KeyGenMock) PublicKeyFromByteArray(_ []byte) (crypto.PublicKey, error) {
	return &PublicKeyMock{}, nil
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
