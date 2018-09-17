package schnorr

import "elrond-go-sandbox/crypto/math"

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

type hash func(string, math.Point) math.Scalar

// x: Private key
func Sign(group math.Group, m string, x math.Scalar, h hash) (math.Point, math.Scalar) {

	g := group.Generator()

	k := group.RandomScalar()

	r := group.Mul(k, g)

	e := h(m, r)

	// s = k - e * x
	s := group.ScalarSub(k, group.ScalarMul(e, x))

	return r, s
}

func PublicKey(group math.Group, m string, r math.Point, s math.Scalar, h hash) math.Point {

	g := group.Generator()

	e := h(m, r)

	// (1 / e) * (r - s * G)
	return group.Mul(group.Inv(e), group.PointSub(r, group.Mul(s, g)))
}

// y: Public key
func Verify(group math.Group, m string, r math.Point, s math.Scalar, y math.Point, h hash) bool {

	g := group.Generator()

	e := h(m, r)

	// s * G = r - e * y
	return group.Equal(group.Mul(s, g), group.PointSub(r, group.Mul(e, y)))
}
