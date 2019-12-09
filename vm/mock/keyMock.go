package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type PublicKeyMock struct {
}

type PrivateKeyMock struct {
}

type KeyGenMock struct {
}

//------- PublicKeyMock

func (sspk *PublicKeyMock) ToByteArray() ([]byte, error) {
	return []byte("pubKey"), nil
}

func (sspk *PublicKeyMock) Suite() crypto.Suite {
	return nil
}

func (sspk *PublicKeyMock) Point() crypto.Point {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sspk *PublicKeyMock) IsInterfaceNil() bool {
	return sspk == nil
}

//------- PrivateKeyMock

func (sk *PrivateKeyMock) ToByteArray() ([]byte, error) {
	return []byte("privKey"), nil
}

func (sk *PrivateKeyMock) GeneratePublic() crypto.PublicKey {
	return &PublicKeyMock{}
}

func (sk *PrivateKeyMock) Suite() crypto.Suite {
	return nil
}

func (sk *PrivateKeyMock) Scalar() crypto.Scalar {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sk *PrivateKeyMock) IsInterfaceNil() bool {
	return sk == nil
}

//------KeyGenMock

func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return &PrivateKeyMock{}, &PublicKeyMock{}
}

func (keyGen *KeyGenMock) PrivateKeyFromByteArray(_ []byte) (crypto.PrivateKey, error) {
	return &PrivateKeyMock{}, nil
}

func (keyGen *KeyGenMock) PublicKeyFromByteArray(_ []byte) (crypto.PublicKey, error) {
	return &PublicKeyMock{}, nil
}

func (keyGen *KeyGenMock) Suite() crypto.Suite {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (keyGen *KeyGenMock) IsInterfaceNil() bool {
	return keyGen == nil
}
