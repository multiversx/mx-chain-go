package mock

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

type SuiteMock struct {
	StringStub             func() string
	ScalarLenStub          func() int
	CreateScalarStub       func() crypto.Scalar
	PointLenStub           func() int
	CreatePointStub        func() crypto.Point
	RandomStreamStub       func() cipher.Stream
	GetUnderlyingSuiteStub func() interface{}
}

type GeneratorSuite struct {
	SuiteMock
	CreateKeyStub func(cipher.Stream) crypto.Scalar
}

func (s *SuiteMock) String() string {
	return s.StringStub()
}

func (s *SuiteMock) ScalarLen() int {
	return s.ScalarLenStub()
}

func (s *SuiteMock) CreateScalar() crypto.Scalar {
	return s.CreateScalarStub()
}

func (s *SuiteMock) PointLen() int {
	return s.PointLenStub()
}

func (s *SuiteMock) CreatePoint() crypto.Point {
	return s.CreatePointStub()
}

func (s *SuiteMock) RandomStream() cipher.Stream {
	stream := NewStreamer()
	return stream
}

func (s *SuiteMock) GetUnderlyingSuite() interface{} {
	return s.GetUnderlyingSuiteStub()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SuiteMock) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}

func (gs *GeneratorSuite) CreateKey(c cipher.Stream) crypto.Scalar {
	return gs.CreateKeyStub(c)
}
