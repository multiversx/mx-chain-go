package bn

import "elrond-go-sandbox/crypto/math"

type signature struct {
	group math.Group
	h     math.Hash
}

func NewSig(group math.Group, h math.Hash) *signature {
	sig := new(signature)
	sig.group = group
	sig.h = h
	return sig
}

func (sig signature) GetL(P []math.Point) math.Scalar {
	s := make([]string, len(P))
	for i := 0; i < len(P); i++ {
		s[i] = sig.group.PointToString(P[i])
	}
	return sig.h(s...)
}

// ei: Private keys
func (sig signature) Sign(g math.Point, ki []math.Scalar, L math.Scalar, Pi []math.Point, m string, ei []math.Scalar) (math.Point, math.Scalar) {

	ri := make([]math.Point, len(ki))
	ci := make([]math.Scalar, len(Pi))
	si := make([]math.Scalar, len(ki))

	for i := 0; i < len(ki); i++ {
		ri[i] = sig.group.Mul(ki[i], g)
	}

	r := sig.sumPoints(ri)

	for i := 0; i < len(Pi); i++ {
		ci[i] = sig.hash(L, Pi[i], r, m)
	}

	for i := 0; i < len(ki); i++ {
		si[i] = sig.group.ScalarAdd(ki[i], sig.group.ScalarMul(ci[i], ei[i]))
	}

	s := sig.sumScalars(si)

	return r, s
}

// Pi: Public keys
func (sig signature) Verify(g math.Point, L math.Scalar, m string, r math.Point, s math.Scalar, Pi []math.Point) bool {

	ci := make([]math.Scalar, len(Pi))
	qi := make([]math.Point, len(Pi))

	for i := 0; i < len(Pi); i++ {
		ci[i] = sig.hash(L, Pi[i], r, m)
	}

	for i := 0; i < len(Pi); i++ {
		qi[i] = sig.group.Mul(ci[i], Pi[i])
	}

	x := sig.group.PointSub(sig.group.Mul(s, g), sig.sumPoints(qi))

	return sig.group.Equal(x, r)
}

func (sig signature) hash(a math.Scalar, b math.Point, c math.Point, d string) math.Scalar {

	return sig.h(sig.group.ScalarToString(a), sig.group.PointToString(b), sig.group.PointToString(c), d)
}

func (sig signature) sumPoints(p []math.Point) math.Point {

	var sum = p[0]

	for i := 1; i < len(p); i++ {
		sum = sig.group.PointAdd(sum, p[i])
	}

	return sum
}

func (sig signature) sumScalars(s []math.Scalar) math.Scalar {

	var sum = s[0]

	for i := 1; i < len(s); i++ {
		sum = sig.group.ScalarAdd(sum, s[i])
	}

	return sum
}
