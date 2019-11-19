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

type KeyGenStub struct {
	GeneratePairCalled            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayCalled func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayCalled  func(b []byte) (crypto.PublicKey, error)
	SuiteCalled                   func() crypto.Suite
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
	if sspk == nil {
		return true
	}
	return false
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
	if sk == nil {
		return true
	}
	return false
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

// IsInterfaceNil returns true if there is no value under the interface
func (keyGen *KeyGenMock) IsInterfaceNil() bool {
	if keyGen == nil {
		return true
	}
	return false
}

func (kgs *KeyGenStub) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return kgs.GeneratePairCalled()
}

func (kgs *KeyGenStub) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	return kgs.PrivateKeyFromByteArrayCalled(b)
}

func (kgs *KeyGenStub) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return kgs.PublicKeyFromByteArrayCalled(b)
}

func (kgs *KeyGenStub) Suite() crypto.Suite {
	return kgs.SuiteCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (kgs *KeyGenStub) IsInterfaceNil() bool {
	return kgs == nil
}
