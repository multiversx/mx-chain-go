package mock

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

// SuiteMock -
type SuiteMock struct {
	StringStub             func() string
	ScalarLenStub          func() int
	CreateScalarStub       func() crypto.Scalar
	PointLenStub           func() int
	CreatePointStub        func() crypto.Point
	RandomStreamStub       func() cipher.Stream
	CreateKeyPairStub      func(cipher.Stream) (crypto.Scalar, crypto.Point)
	GetUnderlyingSuiteStub func() interface{}
	GetPublicKeyPointStub  func(scalar crypto.Scalar) (crypto.Point, error)
}

// String -
func (s *SuiteMock) String() string {
	return s.StringStub()
}

// ScalarLen -
func (s *SuiteMock) ScalarLen() int {
	return s.ScalarLenStub()
}

// CreateScalar -
func (s *SuiteMock) CreateScalar() crypto.Scalar {
	return s.CreateScalarStub()
}

// PointLen -
func (s *SuiteMock) PointLen() int {
	return s.PointLenStub()
}

// CreatePoint -
func (s *SuiteMock) CreatePoint() crypto.Point {
	return s.CreatePointStub()
}

// RandomStream -
func (s *SuiteMock) RandomStream() cipher.Stream {
	stream := NewStreamer()
	return stream
}

// GetUnderlyingSuite -
func (s *SuiteMock) GetUnderlyingSuite() interface{} {
	return s.GetUnderlyingSuiteStub()
}

// CreateKeyPair -
func (s *SuiteMock) CreateKeyPair(c cipher.Stream) (crypto.Scalar, crypto.Point) {
	return s.CreateKeyPairStub(c)
}

// GetPublicKeyPoint -
func (s *SuiteMock) GetPublicKeyPoint(scalar crypto.Scalar) (crypto.Point, error) {
	return  s.GetPublicKeyPointStub(scalar)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SuiteMock) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
