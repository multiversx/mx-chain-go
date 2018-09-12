package schnorr

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

import (
	"fmt"
	"gopkg.in/dedis/kyber.v2"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
)

var curve = edwards25519.NewBlakeSHA256Ed25519()
var sha256 = curve.Hash()

type Signature struct {
	r kyber.Point
	s kyber.Scalar
}

type Group interface {
	G() kyber.Point
	RandomScalar() kyber.Scalar
	Mul(kyber.Scalar, kyber.Point) kyber.Point
	PointSub(a, b kyber.Point) kyber.Point
	ScalarSub(a, b kyber.Scalar) kyber.Scalar
	ScalarMul(scalar kyber.Scalar) kyber.Scalar
	Inv(scalar kyber.Scalar) kyber.Scalar
}

type Ed25519Group struct {
}

func (group Ed25519Group) G() kyber.Point {
	return curve.Point().Base()
}

func (group Ed25519Group) RandomScalar() kyber.Scalar {
	return curve.Scalar().Pick(curve.RandomStream())
}

func (group Ed25519Group) Mul(scalar kyber.Scalar, point kyber.Point) kyber.Point {
	return curve.Point().Mul(scalar, point)
}

func (group Ed25519Group) PointSub(a, b kyber.Point) kyber.Point {
	return curve.Point().Sub(a, b)
}

func (group Ed25519Group) ScalarSub(a, b kyber.Scalar) kyber.Scalar {
	return curve.Scalar().Sub(a, b)
}

func (group Ed25519Group) ScalarMul(a, b kyber.Scalar) kyber.Scalar {
	return curve.Scalar().Mul(a, b)
}

func (group Ed25519Group) Inv(scalar kyber.Scalar) kyber.Scalar {
	return curve.Scalar().Div(curve.Scalar().One(), scalar)
}

func Hash(s string) kyber.Scalar {
	sha256.Reset()
	sha256.Write([]byte(s))

	return curve.Scalar().SetBytes(sha256.Sum(nil))
}

var group = Ed25519Group{}
var g = group.G()

// m: Message
// x: Private key
func Sign(m string, x kyber.Scalar) Signature {

	// Pick a random k from allowed set.
	k := group.RandomScalar()

	// r = k * G (a.k.a the same operation as r = g^k)
	r := group.Mul(k, g)

	// Hash(m || r)
	e := Hash(m + r.String())

	// s = k - e * x
	s := group.ScalarSub(k, group.ScalarMul(e, x))

	return Signature{r: r, s: s}
}

// m: Message
// S: Signature
func PublicKey(m string, S Signature) kyber.Point {

	// e = Hash(m || r)
	e := Hash(m + S.r.String())

	// y = (r - s * G) * (1 / e)
	y := group.PointSub(S.r, group.Mul(S.s, g))
	y = group.Mul(group.Inv(e), y)

	return y
}

// m: Message
// s: Signature
// y: Public key
func Verify(m string, S Signature, y kyber.Point) bool {

	// e = Hash(m || r)
	e := Hash(m + S.r.String())

	// Attempt to reconstruct 's * G' with a provided signature; s * G = r - e * y
	sGv := group.PointSub(S.r, group.Mul(e, y))

	// Construct the actual 's * G'
	sG := group.Mul(S.s, g)

	// Equality check; ensure signature and public key outputs to s * G.
	return sG.Equal(sGv)
}

func (S Signature) String() string {
	return fmt.Sprintf("(r=%s, s=%s)", S.r, S.s)
}
