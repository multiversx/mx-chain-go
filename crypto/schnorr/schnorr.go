package schnorr

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

type Point interface{}

type Scalar interface{}

type Group interface {
	Generator() Point
	RandomScalar() Scalar
	Mul(Scalar, Point) Point
	PointSub(a, b Point) Point
	ScalarSub(a, b Scalar) Scalar
	ScalarMul(a, b Scalar) Scalar
	Inv(scalar Scalar) Scalar
	Equal(a, b Point) bool
}

type hash func(string, Point) Scalar

// x: Private key
func Sign(group Group, m string, x Scalar, h hash) (Point, Scalar) {

	g := group.Generator()

	k := group.RandomScalar()

	r := group.Mul(k, g)

	e := h(m, r)

	// s = k - e * x
	s := group.ScalarSub(k, group.ScalarMul(e, x))

	return r, s
}

func PublicKey(group Group, m string, r Point, s Scalar, h hash) Point {

	g := group.Generator()

	e := h(m, r)

	// (1 / e) * (r - s * G)
	return group.Mul(group.Inv(e), group.PointSub(r, group.Mul(s, g)))
}

// y: Public key
func Verify(group Group, m string, r Point, s Scalar, y Point, h hash) bool {

	g := group.Generator()

	e := h(m, r)

	// s * G = r - e * y
	return group.Equal(group.Mul(s, g), group.PointSub(r, group.Mul(e, y)))
}
