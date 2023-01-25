package mock

import (
	"github.com/multiversx/mx-chain-crypto-go"
)

// SingleSignKeyGenMock -
type SingleSignKeyGenMock struct {
	PublicKeyFromByteArrayCalled func(b []byte) (crypto.PublicKey, error)
	SuiteCalled                  func() crypto.Suite
}

// SingleSignPublicKey -
type SingleSignPublicKey struct {
	SuiteCalled func() crypto.Suite
	PointCalled func() crypto.Point
}

//------- SingleSignKeyGenMock

// GeneratePair -
func (sskgm *SingleSignKeyGenMock) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	panic("implement me")
}

// PrivateKeyFromByteArray -
func (sskgm *SingleSignKeyGenMock) PrivateKeyFromByteArray(_ []byte) (crypto.PrivateKey, error) {
	panic("implement me")
}

// PublicKeyFromByteArray -
func (sskgm *SingleSignKeyGenMock) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return sskgm.PublicKeyFromByteArrayCalled(b)
}

// CheckPublicKeyValid -
func (sskgm *SingleSignKeyGenMock) CheckPublicKeyValid(_ []byte) error {
	return nil
}

// Suite -
func (sskgm *SingleSignKeyGenMock) Suite() crypto.Suite {
	return sskgm.SuiteCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sskgm *SingleSignKeyGenMock) IsInterfaceNil() bool {
	return sskgm == nil
}

//------- SingleSignPublicKey

// ToByteArray -
func (sspk *SingleSignPublicKey) ToByteArray() ([]byte, error) {
	panic("implement me")
}

// Suite -
func (sspk *SingleSignPublicKey) Suite() crypto.Suite {
	return sspk.SuiteCalled()
}

// Point -
func (sspk *SingleSignPublicKey) Point() crypto.Point {
	return sspk.PointCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sspk *SingleSignPublicKey) IsInterfaceNil() bool {
	return sspk == nil
}
