package schnorr

import "github.com/ElrondNetwork/elrond-go-sandbox/crypto/math"

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

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

// x: Private key
func (sig signature) Sign(g math.Point, k math.Scalar, m string, x math.Scalar) (math.Point, math.Scalar) {

	r := sig.group.Mul(k, g)

	e := sig.hash(m, r)

	// s = k - e * x
	s := sig.group.ScalarSub(k, sig.group.ScalarMul(e, x))

	return r, s
}

func (sig signature) PublicKey(g math.Point, m string, r math.Point, s math.Scalar) math.Point {

	e := sig.hash(m, r)

	// (1 / e) * (r - s * G)
	return sig.group.Mul(sig.group.Inv(e), sig.group.PointSub(r, sig.group.Mul(s, g)))
}

// y: Public key
func (sig signature) Verify(g math.Point, m string, r math.Point, s math.Scalar, y math.Point) bool {

	e := sig.hash(m, r)

	// s * G = r - e * y
	return sig.group.Equal(sig.group.Mul(s, g), sig.group.PointSub(r, sig.group.Mul(e, y)))
}

func (sig signature) hash(m string, r math.Point) math.Scalar {
	return sig.h(m, sig.group.PointToString(r))
}
