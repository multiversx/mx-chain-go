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
func (sig signature) Sign(g math.Point, k []math.Scalar, L math.Scalar, P []math.Point, m string, e []math.Scalar) (math.Point, math.Scalar) {

	R := make([]math.Point, len(k))
	c := make([]math.Scalar, len(P))
	S := make([]math.Scalar, len(k))

	for i := 0; i < len(k); i++ {
		R[i] = sig.group.Mul(k[i], g)
	}

	r := sig.sumPoints(R)

	for i := 0; i < len(P); i++ {
		c[i] = sig.hash(L, P[i], r, m)
	}

	for i := 0; i < len(k); i++ {
		S[i] = sig.group.ScalarAdd(k[i], sig.group.ScalarMul(c[i], e[i]))
	}

	s := sig.sumScalars(S)

	return r, s
}

// Pi: Public keys
func (sig signature) Verify(g math.Point, L math.Scalar, m string, r math.Point, s math.Scalar, P []math.Point) bool {

	c := make([]math.Scalar, len(P))
	cP := make([]math.Point, len(P))

	for i := 0; i < len(P); i++ {
		c[i] = sig.hash(L, P[i], r, m)
	}

	for i := 0; i < len(P); i++ {
		cP[i] = sig.group.Mul(c[i], P[i])
	}

	x := sig.group.PointSub(sig.group.Mul(s, g), sig.sumPoints(cP))

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
