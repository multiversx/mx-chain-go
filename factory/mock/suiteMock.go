package mock

import (
	"crypto/cipher"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

// SuiteMock -
type SuiteMock struct {
	StringStub               func() string
	ScalarLenStub            func() int
	CreateScalarStub         func() crypto.Scalar
	PointLenStub             func() int
	CreatePointStub          func() crypto.Point
	CreatePointForScalarStub func(scalar crypto.Scalar) (crypto.Point, error)
	RandomStreamStub         func() cipher.Stream
	CreateKeyPairStub        func() (crypto.Scalar, crypto.Point)
	GetUnderlyingSuiteStub   func() interface{}
}

// String -
func (s *SuiteMock) String() string {
	if s.StringStub != nil {
		return s.StringStub()
	}
	return "mock suite"
}

// ScalarLen -
func (s *SuiteMock) ScalarLen() int {
	if s.ScalarLenStub != nil {
		return s.ScalarLenStub()
	}

	return 32
}

// CreateScalar -
func (s *SuiteMock) CreateScalar() crypto.Scalar {
	if s.CreateScalarStub != nil {
		return s.CreateScalarStub()
	}
	return nil
}

// PointLen -
func (s *SuiteMock) PointLen() int {
	if s.PointLenStub != nil {
		return s.PointLenStub()
	}
	return 64
}

// CreatePoint -
func (s *SuiteMock) CreatePoint() crypto.Point {
	if s.CreatePointStub != nil {
		return s.CreatePointStub()
	}
	return nil
}

// CreatePointForScalar -
func (s *SuiteMock) CreatePointForScalar(scalar crypto.Scalar) (crypto.Point, error) {
	if s.CreatePointForScalarStub != nil {
		return s.CreatePointForScalarStub(scalar)
	}
	return nil, nil
}

// RandomStream -
func (s *SuiteMock) RandomStream() cipher.Stream {
	stream := NewStreamer()
	return stream
}

// GetUnderlyingSuite -
func (s *SuiteMock) GetUnderlyingSuite() interface{} {
	if s.GetUnderlyingSuiteStub != nil {
		return s.GetUnderlyingSuiteStub()
	}
	return "invalid suite"
}

// CreateKeyPair -
func (s *SuiteMock) CreateKeyPair() (crypto.Scalar, crypto.Point) {
	if s.CreateKeyPairStub != nil {
		return s.CreateKeyPairStub()
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SuiteMock) IsInterfaceNil() bool {
	return s == nil
}

// Streamer -
type Streamer struct {
	key []byte
}

// NewStreamer -
func NewStreamer() *Streamer {
	key, _ := hex.DecodeString("aa")
	return &Streamer{key: key}
}

// XORKeyStream -
func (stream *Streamer) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("dst length < src length")
	}

	for i := 0; i < len(src); i++ {
		dst[i] = src[i] ^ stream.key[0]
	}
}
