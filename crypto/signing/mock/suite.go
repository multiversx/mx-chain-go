package mock

import (
	"crypto/cipher"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type Suite struct {
	StringStub             func() string
	ScalarLenStub          func() int
	CreateScalarStub       func() crypto.Scalar
	PointLenStub           func() int
	CreatePointStub        func() crypto.Point
	RandomStreamStub       func() cipher.Stream
	GetUnderlyingSuiteStub func() interface{}
}

func (s *Suite) String() string {
	return s.StringStub()
}

func (s *Suite) ScalarLen() int {
	return s.ScalarLenStub()
}

func (s *Suite) CreateScalar() crypto.Scalar {
	return s.CreateScalarStub()
}

func (s *Suite) PointLen() int {
	return s.PointLenStub()
}

func (s *Suite) CreatePoint() crypto.Point {
	return s.CreatePointStub()
}

func (s *Suite) RandomStream() cipher.Stream {
	panic("implement me")
}

func (s *Suite) GetUnderlyingSuite() interface{} {
	return s.GetUnderlyingSuiteStub()
}
