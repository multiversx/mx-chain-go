package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type SingleSignKeyGenMock struct {
	PublicKeyFromByteArrayCalled func(b []byte) (crypto.PublicKey, error)
	SuiteCalled                  func() crypto.Suite
}

type SingleSignPublicKeyMock struct {
	VerifyCalled func(data []byte, signature []byte, signer crypto.SingleSigner) error
	SuiteCalled  func() crypto.Suite
	PointCalled  func() crypto.Point
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

func (sskgm *SingleSignKeyGenMock) Suite() crypto.Suite {
	return sskgm.SuiteCalled()
}

//------- SingleSignPublicKey

func (sspk *SingleSignPublicKeyMock) ToByteArray() ([]byte, error) {
	panic("implement me")
}

func (sspk *SingleSignPublicKeyMock) Verify(data []byte, signature []byte, signer crypto.SingleSigner) error {
	return sspk.VerifyCalled(data, signature, signer)
}

func (sspk *SingleSignPublicKeyMock) Suite() crypto.Suite {
	return sspk.SuiteCalled()
}

func (sspk *SingleSignPublicKeyMock) Point() crypto.Point {
	return sspk.PointCalled()
}
