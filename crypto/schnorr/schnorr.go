package schnorr

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

import (
	"gopkg.in/dedis/kyber.v2"
)

type Group interface {
	Generator() kyber.Point
	RandomScalar() kyber.Scalar
	Mul(kyber.Scalar, kyber.Point) kyber.Point
	PointSub(a, b kyber.Point) kyber.Point
	ScalarSub(a, b kyber.Scalar) kyber.Scalar
	ScalarMul(a, b kyber.Scalar) kyber.Scalar
	Inv(scalar kyber.Scalar) kyber.Scalar
}

type hash func(string, kyber.Point) kyber.Scalar

// x: Private key
func Sign(group Group, m string, x kyber.Scalar, h hash) (kyber.Point, kyber.Scalar) {

	g := group.Generator()

	k := group.RandomScalar()

	r := group.Mul(k, g)

	e := h(m, r)

	// s = k - e * x
	s := group.ScalarSub(k, group.ScalarMul(e, x))

	return r, s
}

func PublicKey(group Group, m string, r kyber.Point, s kyber.Scalar, h hash) kyber.Point {

	g := group.Generator()

	e := h(m, r)

	// (1 / e) * (r - s * G)
	return group.Mul(group.Inv(e), group.PointSub(r, group.Mul(s, g)))
}

// y: Public key
func Verify(group Group, m string, r kyber.Point, s kyber.Scalar, y kyber.Point, h func(string, kyber.Point) kyber.Scalar) bool {

	g := group.Generator()

	e := h(m, r)

	// s * G = r - e * y
	return group.Mul(s, g).Equal(group.PointSub(r, group.Mul(e, y)))
}
