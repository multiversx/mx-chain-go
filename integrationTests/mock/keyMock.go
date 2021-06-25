package mock

import (
	"crypto/rand"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
)

// PublicKeyMock -
type PublicKeyMock struct {
	pubKey []byte
}

// PrivateKeyMock -
type PrivateKeyMock struct {
	privKey []byte
}

// KeyGenMock -
type KeyGenMock struct {
}

// ToByteArray -
func (sspk *PublicKeyMock) ToByteArray() ([]byte, error) {
	return sspk.pubKey, nil
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
	return sk.privKey, nil
}

// GeneratePublic -
func (sk *PrivateKeyMock) GeneratePublic() crypto.PublicKey {
	return &PublicKeyMock{
		pubKey: sha256.NewSha256().Compute(string(sk.privKey)),
	}
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
	buff := make([]byte, 32)
	_, _ = rand.Read(buff)

	sk := &PrivateKeyMock{
		privKey: buff,
	}

	return sk, sk.GeneratePublic()
}

// PrivateKeyFromByteArray -
func (keyGen *KeyGenMock) PrivateKeyFromByteArray(buff []byte) (crypto.PrivateKey, error) {
	return &PrivateKeyMock{
		privKey: buff,
	}, nil
}

// PublicKeyFromByteArray -
func (keyGen *KeyGenMock) PublicKeyFromByteArray(buff []byte) (crypto.PublicKey, error) {
	return &PublicKeyMock{
		pubKey: buff,
	}, nil
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
