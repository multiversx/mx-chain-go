package ed25519

import (
	"elrond-go-sandbox/crypto/math"
	"gopkg.in/dedis/kyber.v2"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
)

var curve = edwards25519.NewBlakeSHA256Ed25519()

type Group struct {
}

func (group Group) Generator() math.Point {
	return curve.Point().Base()
}

func (group Group) RandomScalar() math.Scalar {
	return curve.Scalar().Pick(curve.RandomStream())
}

func (group Group) Mul(scalar math.Scalar, point math.Point) math.Point {
	return curve.Point().Mul(scalar.(kyber.Scalar), point.(kyber.Point))
}

func (group Group) PointSub(a, b math.Point) math.Point {
	return curve.Point().Sub(a.(kyber.Point), b.(kyber.Point))
}

func (group Group) ScalarSub(a, b math.Scalar) math.Scalar {
	return curve.Scalar().Sub(a.(kyber.Scalar), b.(kyber.Scalar))
}

func (group Group) ScalarMul(a, b math.Scalar) math.Scalar {
	return curve.Scalar().Mul(a.(kyber.Scalar), b.(kyber.Scalar))
}

func (group Group) Inv(scalar math.Scalar) math.Scalar {
	return curve.Scalar().Div(curve.Scalar().One(), scalar.(kyber.Scalar))
}

func (group Group) Equal(a, b math.Point) bool {
	return a.(kyber.Point).Equal(b.(kyber.Point))
}

var sha256 = curve.Hash()

func Hash(s string, point math.Point) math.Scalar {
	sha256.Reset()
	sha256.Write([]byte(s + point.(kyber.Point).String()))

	return curve.Scalar().SetBytes(sha256.Sum(nil))
}
