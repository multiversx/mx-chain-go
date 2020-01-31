package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// PublicKeyMock -
type PublicKeyMock struct {
	ToByteArrayHandler func() ([]byte, error)
	SuiteCalled        func() crypto.Suite
	PointCalled        func() crypto.Point
}

// PrivateKeyStub -
type PrivateKeyStub struct {
	ToByteArrayHandler    func() ([]byte, error)
	GeneratePublicHandler func() crypto.PublicKey
	SuiteHandler          func() crypto.Suite
	ScalarHandler         func() crypto.Scalar
}

// KeyGenMock -
type KeyGenMock struct {
	GeneratePairMock            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayMock func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayMock  func(b []byte) (crypto.PublicKey, error)
	SuiteMock                   func() crypto.Suite
}

// ToByteArray -
func (sspk *PublicKeyMock) ToByteArray() ([]byte, error) {
	return sspk.ToByteArrayHandler()
}

// Suite -
func (sspk *PublicKeyMock) Suite() crypto.Suite {
	return sspk.SuiteCalled()
}

// Point -
func (sspk *PublicKeyMock) Point() crypto.Point {
	return sspk.PointCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sspk *PublicKeyMock) IsInterfaceNil() bool {
	if sspk == nil {
		return true
	}
	return false
}

// ToByteArray -
func (sk *PrivateKeyStub) ToByteArray() ([]byte, error) {
	return sk.ToByteArrayHandler()
}

// GeneratePublic -
func (sk *PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	return sk.GeneratePublicHandler()
}

// Suite -
func (sk *PrivateKeyStub) Suite() crypto.Suite {
	return sk.SuiteHandler()
}

// Scalar -
func (sk *PrivateKeyStub) Scalar() crypto.Scalar {
	return sk.ScalarHandler()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sk *PrivateKeyStub) IsInterfaceNil() bool {
	if sk == nil {
		return true
	}
	return false
}

// GeneratePair -
func (keyGen *KeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return &mock.PrivateKeyMock{}, &mock.PublicKeyMock{}
}

// PrivateKeyFromByteArray -
func (keyGen *KeyGenMock) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	return keyGen.PrivateKeyFromByteArrayMock(b)
}

// PublicKeyFromByteArray -
func (keyGen *KeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return keyGen.PublicKeyFromByteArrayMock(b)
}

// Suite -
func (keyGen *KeyGenMock) Suite() crypto.Suite {
	return keyGen.SuiteMock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (keyGen *KeyGenMock) IsInterfaceNil() bool {
	if keyGen == nil {
		return true
	}
	return false
}
