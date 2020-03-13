package mcl

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

type groupG1 struct {
}

// String returns the string for the group
func (g1 *groupG1) String() string {
	return "BLS12-381 G1"
}

// ScalarLen returns the maximum length of scalars in bytes
func (g1 *groupG1) ScalarLen() int {
	return bls.GetFrByteSize()
}

// CreateScalar creates a new Scalar initialized with base point on G1
func (g1 *groupG1) CreateScalar() crypto.Scalar {
	return NewScalar()
}

// PointLen returns the max length of point in nb of bytes
func (g1 *groupG1) PointLen() int {
	return bls.GetG1ByteSize()
}

// CreatePoint creates a new point
func (g1 *groupG1) CreatePoint() crypto.Point {
	return NewPointG1()
}

// CreatePointForScalar creates a new point corresponding to the given scalarInt
func (g1 *groupG1) CreatePointForScalar(scalar crypto.Scalar) crypto.Point {
	var p crypto.Point
	var err error
	p = NewPointG1()
	p, err = p.Mul(scalar)
	if err != nil {
		log.Error("groupG1 CreatePointForScalar", "error", err.Error())
	}
	return p
}

// IsInterfaceNil returns true if there is no value under the interface
func (g1 *groupG1) IsInterfaceNil() bool {
	return g1 == nil
}
