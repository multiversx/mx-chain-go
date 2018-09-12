package ed25519

import (
	"gopkg.in/dedis/kyber.v2"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
)

var curve = edwards25519.NewBlakeSHA256Ed25519()

type Group struct {
}

func (group Group) G() kyber.Point {
	return curve.Point().Base()
}

func (group Group) RandomScalar() kyber.Scalar {
	return curve.Scalar().Pick(curve.RandomStream())
}

func (group Group) Mul(scalar kyber.Scalar, point kyber.Point) kyber.Point {
	return curve.Point().Mul(scalar, point)
}

func (group Group) PointSub(a, b kyber.Point) kyber.Point {
	return curve.Point().Sub(a, b)
}

func (group Group) ScalarSub(a, b kyber.Scalar) kyber.Scalar {
	return curve.Scalar().Sub(a, b)
}

func (group Group) ScalarMul(a, b kyber.Scalar) kyber.Scalar {
	return curve.Scalar().Mul(a, b)
}

func (group Group) Inv(scalar kyber.Scalar) kyber.Scalar {
	return curve.Scalar().Div(curve.Scalar().One(), scalar)
}
