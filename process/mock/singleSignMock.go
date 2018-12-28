package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type SingleSignKeyGenMock struct {
	PublicKeyFromByteArrayCalled func(b []byte) (crypto.PublicKey, error)
}

type SingleSignPublicKey struct {
	VerifyCalled func(data []byte, signature []byte) (bool, error)
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

func (sspk *SingleSignPublicKey) ToByteArray() ([]byte, error) {
	panic("implement me")
}

func (sspk *SingleSignPublicKey) Verify(data []byte, signature []byte) (bool, error) {
	return sspk.VerifyCalled(data, signature)
}
