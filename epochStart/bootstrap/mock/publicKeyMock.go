package mock

import "github.com/ElrondNetwork/elrond-go/crypto"

// PublicKeyMock mocks a public key implementation
type PublicKeyMock struct {
	ToByteArrayMock func() ([]byte, error)
	SuiteMock       func() crypto.Suite
	PointMock       func() crypto.Point
}

// ToByteArray mocks converting a public key to a byte array
func (pubKey *PublicKeyMock) ToByteArray() ([]byte, error) {
	return []byte("publicKeyMock"), nil
}

// Suite -
func (pubKey *PublicKeyMock) Suite() crypto.Suite {
	return pubKey.SuiteMock()
}

// Point -
func (pubKey *PublicKeyMock) Point() crypto.Point {
	return pubKey.PointMock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pubKey *PublicKeyMock) IsInterfaceNil() bool {
	return pubKey == nil
}
