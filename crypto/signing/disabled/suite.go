package disabled

import (
	"crypto/cipher"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

// Disabled is the string for a disabled suite
const Disabled = "disabled"
const scalarLen = 32
const pointLen = 96

type disabledSuite struct{}

// NewDisabledSuite creates a new disabled suite
func NewDisabledSuite() *disabledSuite {
	return &disabledSuite{}
}

// String returns the disabled string
func (ds *disabledSuite) String() string {
	return Disabled
}

// ScalarLen returns 0
func (ds *disabledSuite) ScalarLen() int {
	return scalarLen
}

// CreateScalar returns a disabledScalar instance
func (ds *disabledSuite) CreateScalar() crypto.Scalar {
	return &disabledScalar{}
}

// PointLen returns 0
func (ds *disabledSuite) PointLen() int {
	return pointLen
}

// CreatePoint creates a disabledPoint instance
func (ds *disabledSuite) CreatePoint() crypto.Point {
	return &disabledPoint{}
}

// CreatePointForScalar creates a disabledPoint instance
func (ds *disabledSuite) CreatePointForScalar(_ crypto.Scalar) (crypto.Point, error) {
	return &disabledPoint{}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ds *disabledSuite) IsInterfaceNil() bool {
	return ds == nil
}

// RandomStream returns nil
func (ds *disabledSuite) RandomStream() cipher.Stream {
	return nil
}

// CreateKeyPair returns a disabled scalar and point pair
func (ds *disabledSuite) CreateKeyPair() (crypto.Scalar, crypto.Point) {
	return &disabledScalar{}, &disabledPoint{}
}

// CheckPointValid returns nil
func (ds *disabledSuite) CheckPointValid(_ []byte) error {
	return nil
}

// GetUnderlyingSuite returns nil
func (ds *disabledSuite) GetUnderlyingSuite() interface{} {
	return nil
}
