package mock

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
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
	CreateKeyStub func(cipher.Stream) (crypto.Scalar, crypto.Point)
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

func (gs *GeneratorSuite) CreateKeyPair(c cipher.Stream) (crypto.Scalar, crypto.Point) {
	return gs.CreateKeyStub(c)
}
