package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type PublicKeyMock struct {
	ToByteArrayHandler func() ([]byte, error)
	SuiteCalled        func() crypto.Suite
	PointCalled        func() crypto.Point
}

type PrivateKeyStub struct {
	ToByteArrayHandler    func() ([]byte, error)
	GeneratePublicHandler func() crypto.PublicKey
	SuiteHandler          func() crypto.Suite
	ScalarHandler         func() crypto.Scalar
}

type KeyGenMock struct {
	GeneratePairMock            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayMock func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayMock  func(b []byte) (crypto.PublicKey, error)
	SuiteMock                   func() crypto.Suite
}

//------- PublicKeyMock

func (sspk *PublicKeyMock) ToByteArray() ([]byte, error) {
	return sspk.ToByteArrayHandler()
}

func (sspk *PublicKeyMock) Suite() crypto.Suite {
	return sspk.SuiteCalled()
}

func (sspk *PublicKeyMock) Point() crypto.Point {
	return sspk.PointCalled()
}

//------- PrivateKeyMock

func (sk *PrivateKeyStub) ToByteArray() ([]byte, error) {
	return sk.ToByteArrayHandler()
}

func (sk *PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	return sk.GeneratePublicHandler()
}

func (sk *PrivateKeyStub) Suite() crypto.Suite {
	return sk.SuiteHandler()
}

func (sk *PrivateKeyStub) Scalar() crypto.Scalar {
	return sk.ScalarHandler()
}

//------KeyGenMock

func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return &mock.PrivateKeyMock{}, &mock.PublicKeyMock{}
}

func (keyGen *KeyGenMock) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	return keyGen.PrivateKeyFromByteArrayMock(b)
}

func (keyGen *KeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return keyGen.PublicKeyFromByteArrayMock(b)
}

func (keyGen *KeyGenMock) Suite() crypto.Suite {
	return keyGen.SuiteMock()
}
