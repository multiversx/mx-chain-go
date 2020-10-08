package disabled

import "github.com/ElrondNetwork/elrond-go/crypto"

const marshaledPoint = "point"

type disabledPoint struct {
}

// MarshalBinary returns a mock value
func (dp *disabledPoint) MarshalBinary() ([]byte, error) {
	return []byte(marshaledPoint), nil
}

// UnmarshalBinary returns nil
func (dp *disabledPoint) UnmarshalBinary(_ []byte) error {
	return nil
}

// Equal returns false
func (dp *disabledPoint) Equal(_ crypto.Point) (bool, error) {
	return false, nil
}

// Null returns a new disabledPoint instance
func (dp *disabledPoint) Null() crypto.Point {
	return &disabledPoint{}
}

// Set returns nil
func (dp *disabledPoint) Set(_ crypto.Point) error {
	return nil
}

// Clone returns a new disabledPoint instance
func (dp *disabledPoint) Clone() crypto.Point {
	return &disabledPoint{}
}

// Add returns a new disabledPoint instance
func (dp *disabledPoint) Add(_ crypto.Point) (crypto.Point, error) {
	return &disabledPoint{}, nil
}

// Sub returns a new disabledPoint instance
func (dp *disabledPoint) Sub(_ crypto.Point) (crypto.Point, error) {
	return &disabledPoint{}, nil
}

// Neg returns a new disabledPoint instance
func (dp *disabledPoint) Neg() crypto.Point {
	return &disabledPoint{}
}

// Mul returns a new disabledPoint instance
func (dp *disabledPoint) Mul(_ crypto.Scalar) (crypto.Point, error) {
	return &disabledPoint{}, nil
}

// Pick returns a new disabledPoint instance
func (dp *disabledPoint) Pick() (crypto.Point, error) {
	return &disabledPoint{}, nil
}

// GetUnderlyingObj returns nil
func (dp *disabledPoint) GetUnderlyingObj() interface{} {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dp *disabledPoint) IsInterfaceNil() bool {
	return dp == nil
}
