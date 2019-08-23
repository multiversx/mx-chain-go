package mock

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type SingleSignKeyGenMock struct {
	PublicKeyFromByteArrayCalled func(b []byte) (crypto.PublicKey, error)
	SuiteCalled                  func() crypto.Suite
}

type SingleSignPublicKey struct {
	SuiteCalled func() crypto.Suite
	PointCalled func() crypto.Point
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

// IsInterfaceNil returns true if there is no value under the interface
func (sskgm *SingleSignKeyGenMock) IsInterfaceNil() bool {
	if sskgm == nil {
		return true
	}
	return false
}

//------- SingleSignPublicKey

func (sspk *SingleSignPublicKey) ToByteArray() ([]byte, error) {
	panic("implement me")
}

func (sspk *SingleSignPublicKey) Suite() crypto.Suite {
	return sspk.SuiteCalled()
}

func (sspk *SingleSignPublicKey) Point() crypto.Point {
	return sspk.PointCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sspk *SingleSignPublicKey) IsInterfaceNil() bool {
	if sspk == nil {
		return true
	}
	return false
}
