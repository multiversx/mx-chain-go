package mock

import (
	"crypto/cipher"

	"github.com/multiversx/mx-chain-crypto-go"
)

// SuiteMock -
type SuiteMock struct {
	StringStub               func() string
	ScalarLenStub            func() int
	CreateScalarStub         func() crypto.Scalar
	PointLenStub             func() int
	CreatePointStub          func() crypto.Point
	CreatePointForScalarStub func(scalar crypto.Scalar) crypto.Point
	RandomStreamStub         func() cipher.Stream
	GetUnderlyingSuiteStub   func() interface{}
}

// GeneratorSuite -
type GeneratorSuite struct {
	SuiteMock
	CreateKeyStub func(cipher.Stream) crypto.Scalar
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

// IsInterfaceNil returns true if there is no value under the interface
func (s *SuiteMock) IsInterfaceNil() bool {
	return s == nil
}

// CreateKey -
func (gs *GeneratorSuite) CreateKey(c cipher.Stream) crypto.Scalar {
	return gs.CreateKeyStub(c)
}
