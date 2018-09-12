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
	Mul(kyber.Scalar, kyber.Point) kyber.Point
}

type Ed25519Group struct {
}

func (group Ed25519Group) G() kyber.Point {
	return curve.Point().Base()
}

func (group Ed25519Group) Mul(scalar kyber.Scalar, point kyber.Point) kyber.Point {
	return curve.Point().Mul(scalar, point)
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
	k := curve.Scalar().Pick(curve.RandomStream())

	// r = k * G (a.k.a the same operation as r = g^k)
	r := group.Mul(k, g)

	// Hash(m || r)
	e := Hash(m + r.String())

	// s = k - e * x
	s := curve.Scalar().Sub(k, curve.Scalar().Mul(e, x))

	return Signature{r: r, s: s}
}

// m: Message
// S: Signature
func PublicKey(m string, S Signature) kyber.Point {

	// e = Hash(m || r)
	e := Hash(m + S.r.String())

	// y = (r - s * G) * (1 / e)
	y := curve.Point().Sub(S.r, group.Mul(S.s, g))
	y = curve.Point().Mul(curve.Scalar().Div(curve.Scalar().One(), e), y)

	return y
}

// m: Message
// s: Signature
// y: Public key
func Verify(m string, S Signature, y kyber.Point) bool {
	// Create a generator.
	g := curve.Point().Base()

	// e = Hash(m || r)
	e := Hash(m + S.r.String())

	// Attempt to reconstruct 's * G' with a provided signature; s * G = r - e * y
	sGv := curve.Point().Sub(S.r, group.Mul(e, y))

	// Construct the actual 's * G'
	sG := group.Mul(S.s, g)

	// Equality check; ensure signature and public key outputs to s * G.
	return sG.Equal(sGv)
}

func (S Signature) String() string {
	return fmt.Sprintf("(r=%s, s=%s)", S.r, S.s)
}
