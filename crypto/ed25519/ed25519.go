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

func (group Group) PointAdd(a, b math.Point) math.Point {
	return curve.Point().Add(a.(kyber.Point), b.(kyber.Point))
}

func (group Group) PointSub(a, b math.Point) math.Point {
	return curve.Point().Sub(a.(kyber.Point), b.(kyber.Point))
}

func (group Group) ScalarAdd(a, b math.Scalar) math.Scalar {
	return curve.Scalar().Add(a.(kyber.Scalar), b.(kyber.Scalar))
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

func HashBN(a math.Scalar, b math.Point, c math.Point, d string) math.Scalar {
	sha256.Reset()
	sha256.Write([]byte(a.(kyber.Scalar).String() + b.(kyber.Point).String() + c.(kyber.Point).String() + d))

	return curve.Scalar().SetBytes(sha256.Sum(nil))
}

func HashBNPublicKeys(p []math.Point) math.Scalar {
	sha256.Reset()

	result := ""
	for i := 0; i < len(p); i++ {
		result += p[i].(kyber.Point).String()
	}
	sha256.Write([]byte(result))

	return curve.Scalar().SetBytes(sha256.Sum(nil))
}
