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
