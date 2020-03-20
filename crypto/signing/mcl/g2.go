package mcl

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

type groupG2 struct {
}

// String returns the string for the group
func (g2 *groupG2) String() string {
	return "BLS12-381 G2"
}

// ScalarLen returns the maximum length of scalars in bytes
func (g2 *groupG2) ScalarLen() int {
	return bls.GetFrByteSize()
}

// CreateScalar creates a new Scalar
func (g2 *groupG2) CreateScalar() crypto.Scalar {
	return NewScalar()
}

// PointLen returns the max length of point in nb of bytes
func (g2 *groupG2) PointLen() int {
	return bls.GetG2ByteSize()
}

// CreatePoint creates a new point
func (g2 *groupG2) CreatePoint() crypto.Point {
	return NewPointG2()
}

// CreatePointForScalar creates a new point corresponding to the given scalarInt
func (g2 *groupG2) CreatePointForScalar(scalar crypto.Scalar) crypto.Point {
	var p crypto.Point
	var err error
	p = NewPointG2()
	p, err = p.Mul(scalar)
	if err != nil {
		log.Error("groupG2 CreatePointForScalar", "error", err.Error())
	}
	return p
}

// IsInterfaceNil returns true if there is no value under the interface
func (g2 *groupG2) IsInterfaceNil() bool {
	return g2 == nil
}
