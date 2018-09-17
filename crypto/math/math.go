package math

type Point interface{}

type Scalar interface{}

type Group interface {
	Mul(Scalar, Point) Point
	PointSub(a, b Point) Point
	ScalarSub(a, b Scalar) Scalar
	ScalarMul(a, b Scalar) Scalar
	Inv(scalar Scalar) Scalar
	Equal(a, b Point) bool
}
