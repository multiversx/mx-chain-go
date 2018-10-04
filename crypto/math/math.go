package math

type Point interface{}

type Scalar interface{}

type Group interface {
	Mul(Scalar, Point) Point
	PointAdd(a, b Point) Point
	PointSub(a, b Point) Point
	ScalarAdd(a, b Scalar) Scalar
	ScalarSub(a, b Scalar) Scalar
	ScalarMul(a, b Scalar) Scalar
	Inv(scalar Scalar) Scalar
	Equal(a, b Point) bool
	ScalarToString(a Scalar) string
	PointToString(a Point) string
}

type Hash func(s ...string) Scalar

func AddPoints(g Group, p []Point) Point {

	var sum = p[0]

	for i := 1; i < len(p); i++ {
		sum = g.PointAdd(sum, p[i])
	}

	return sum
}

func AddScalars(g Group, s []Scalar) Scalar {

	var sum = s[0]

	for i := 1; i < len(s); i++ {
		sum = g.ScalarAdd(sum, s[i])
	}

	return sum
}
