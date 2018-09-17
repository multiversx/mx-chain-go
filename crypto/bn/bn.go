package bn

import "elrond-go-sandbox/crypto/math"

type hash func(math.Scalar, math.Point, math.Point, string) math.Scalar

// ei: Private keys
func Sign(group math.Group, g math.Point, k []math.Scalar, L math.Scalar, Pi []math.Point, m string, ei []math.Scalar, h hash) (math.Point, math.Scalar) {

	ri := mulRange(k, g, group)

	r := sumPoints(ri, group)

	ci := hRange(L, Pi, r, m, h)

	si := sumScalarRange(k, ci, ei, group)

	s := sumScalar(si, group)

	return r, s
}

// Pi: Public keys
func Verify(group math.Group, g math.Point, L math.Scalar, m string, r math.Point, s math.Scalar, Pi []math.Point, h hash) bool {

	ci := hRange(L, Pi, r, m, h)

	x := group.PointSub(group.Mul(s, g), sumPoints(mulRange(ci, Pi, group), group))

	return group.Equal(x, r)
}

func sumPoints(p []math.Point, group math.Group) math.Point {

	var sum = p[0]

	for i := 1; i < len(p); i++ {
		sum = group.PointAdd(sum, p[i])
	}

	return sum
}

func sumScalarRange(k []math.Scalar, c []math.Scalar, e []math.Scalar, group math.Group) []math.Scalar {

	s := make([]math.Scalar, len(k))

	for i := 0; i < len(k); i++ {
		s[i] = group.ScalarAdd(k[i], group.ScalarMul(c[i], e[i]))
	}

	return s
}

func sumScalar(s []math.Scalar, group math.Group) math.Scalar {

	var sum = s[0]

	for i := 1; i < len(s); i++ {
		sum = group.ScalarAdd(sum, s[i])
	}

	return sum
}

func mulRange(k []math.Scalar, g math.Point, group math.Group) []math.Point {

	r := make([]math.Point, len(k))

	for i := 0; i < len(k); i++ {
		r[i] = group.Mul(k[i], g)
	}

	return r
}

func hRange(L math.Scalar, P []math.Point, r math.Point, m string, h hash) []math.Scalar {

	c := make([]math.Scalar, len(P))

	for i := 0; i < len(P); i++ {
		c[i] = h(L, P[i], r, m)
	}

	return c
}
