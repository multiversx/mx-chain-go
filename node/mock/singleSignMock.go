package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type SingleSignKeyGenMock struct {
	PublicKeyFromByteArrayCalled func(b []byte) (crypto.PublicKey, error)
}

type SingleSignPublicKeyMock struct {
	VerifyCalled func(data []byte, signature []byte) error
}

//------- SingleSignKeyGenMock

func (sskgm *SingleSignKeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	panic("implement me")
}

func (sskgm *SingleSignKeyGenMock) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	panic("implement me")
}

func (sskgm *SingleSignKeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return sskgm.PublicKeyFromByteArrayCalled(b)
}

//------- SingleSignPublicKey

func (sspk *SingleSignPublicKeyMock) ToByteArray() ([]byte, error) {
	panic("implement me")
}

func (sspk *SingleSignPublicKeyMock) Verify(data []byte, signature []byte) error {
	return sspk.VerifyCalled(data, signature)
}
