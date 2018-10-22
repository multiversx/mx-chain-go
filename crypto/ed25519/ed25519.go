package ed25519

import (
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/math"
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

func (group Group) PointToString(a math.Point) string {
	return a.(kyber.Point).String()
}

func (group Group) ScalarToString(a math.Scalar) string {
	return a.(kyber.Scalar).String()
}

var sha256 = curve.Hash()

func Hash(s ...string) math.Scalar {
	sha256.Reset()
	sha256.Write([]byte(strings.Join(s, "")))
	return curve.Scalar().SetBytes(sha256.Sum(nil))
}
