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

//------KeyGenMock

func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return &PrivateKeyMock{}, &PublicKeyMock{}
}

func (keyGen *KeyGenMock) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	return &PrivateKeyMock{}, nil
}

func (keyGen *KeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return &PublicKeyMock{}, nil
}

func (keyGen *KeyGenMock) Suite() crypto.Suite {
	return nil
}
