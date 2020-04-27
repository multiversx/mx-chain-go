package mcl

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

type groupGT struct {
}

// String returns the string for the group
func (gt *groupGT) String() string {
	return "BLS12-381 GT"
}

// ScalarLen returns the maximum length of scalars in bytes
func (gt *groupGT) ScalarLen() int {
	return bls.GetFrByteSize()
}

// CreateScalar creates a new Scalar
func (gt *groupGT) CreateScalar() crypto.Scalar {
	return NewScalar()
}

// PointLen returns the max length of point in nb of bytes
func (gt *groupGT) PointLen() int {
	return bls.GetG1ByteSize() * 12
}

// CreatePoint creates a new point
func (gt *groupGT) CreatePoint() crypto.Point {
	return NewPointGT()
}

// CreatePointForScalar creates a new point corresponding to the given scalarInt
func (gt *groupGT) CreatePointForScalar(_ crypto.Scalar) crypto.Point {
	panic("not supported")
}

// IsInterfaceNil returns true if there is no value under the interface
func (gt *groupGT) IsInterfaceNil() bool {
	return gt == nil
}
