package bn

import "elrond-go-sandbox/crypto/math"

type hash func(math.Scalar, math.Point, math.Point, string) math.Scalar

type signature struct {
	group math.Group
	h     hash
}

func NewSig(group math.Group, h hash) *signature {
	sig := new(signature)
	sig.group = group
	sig.h = h
	return sig
}

// ei: Private keys
func (sig signature) Sign(g math.Point, ki []math.Scalar, L math.Scalar, Pi []math.Point, m string, ei []math.Scalar) (math.Point, math.Scalar) {

	ri := sig.mulRange(ki, g)

	r := sig.sumPoints(ri)

	ci := sig.hRange(L, Pi, r, m)

	si := sig.sumScalarRange(ki, ci, ei)

	s := sig.sumScalar(si)

	return r, s
}

// Pi: Public keys
func (sig signature) Verify(g math.Point, L math.Scalar, m string, r math.Point, s math.Scalar, Pi []math.Point) bool {

	ci := sig.hRange(L, Pi, r, m)

	x := sig.group.PointSub(sig.group.Mul(s, g), sig.sumPoints(sig.mulRangeManyPoints(ci, Pi)))

	return sig.group.Equal(x, r)
}

func (sig signature) sumPoints(p []math.Point) math.Point {

	var sum = p[0]

	for i := 1; i < len(p); i++ {
		sum = sig.group.PointAdd(sum, p[i])
	}

	return sum
}

func (sig signature) sumScalarRange(k []math.Scalar, c []math.Scalar, e []math.Scalar) []math.Scalar {

	s := make([]math.Scalar, len(k))

	for i := 0; i < len(k); i++ {
		s[i] = sig.group.ScalarAdd(k[i], sig.group.ScalarMul(c[i], e[i]))
	}

	return s
}

func (sig signature) sumScalar(s []math.Scalar) math.Scalar {

	var sum = s[0]

	for i := 1; i < len(s); i++ {
		sum = sig.group.ScalarAdd(sum, s[i])
	}

	return sum
}

func (sig signature) mulRange(k []math.Scalar, g math.Point) []math.Point {

	r := make([]math.Point, len(k))

	for i := 0; i < len(k); i++ {
		r[i] = sig.group.Mul(k[i], g)
	}

	return r
}

func (sig signature) mulRangeManyPoints(s []math.Scalar, p []math.Point) []math.Point {

	r := make([]math.Point, len(s))

	for i := 0; i < len(s); i++ {
		r[i] = sig.group.Mul(s[i], p[i])
	}

	return r
}

func (sig signature) hRange(L math.Scalar, P []math.Point, r math.Point, m string) []math.Scalar {

	c := make([]math.Scalar, len(P))

	for i := 0; i < len(P); i++ {
		c[i] = sig.h(L, P[i], r, m)
	}

	return c
}
