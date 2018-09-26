package bn_test

import (
	"elrond-go-sandbox/crypto/bn"
	"elrond-go-sandbox/crypto/ed25519"
	"elrond-go-sandbox/crypto/math"
	"fmt"
	"testing"
)

var group = ed25519.Group{}
var hash = ed25519.Hash
var sig = bn.NewSig(group, hash)

func makeRandomScalars(n int) []math.Scalar {

	var k = make([]math.Scalar, n)

	for i := 0; i < len(k); i++ {
		k[i] = group.RandomScalar()
	}

	return k
}

func mulRange(k []math.Scalar, g math.Point) []math.Point {

	var p = make([]math.Point, len(k))

	for i := 0; i < len(k); i++ {
		p[i] = group.Mul(k[i], g)
	}

	return p
}

func hashPublicKeys(p []math.Point) math.Scalar {
	s := make([]string, len(p))
	for i := 0; i < len(p); i++ {
		s[i] = group.PointToString(p[i])
	}
	return hash(s...)
}

func TestSignVerify(t *testing.T) {

	signersCount := 3

	g := group.Generator()

	privateKeys := makeRandomScalars(signersCount)

	publicKeys := mulRange(privateKeys, g)

	fmt.Printf("Generated private keys: %v\n", privateKeys)
	fmt.Printf("Derived public keys: %v\n\n", publicKeys)

	message := "We're gonna be signing this!"

	k := makeRandomScalars(signersCount)

	L := hashPublicKeys(publicKeys)

	r, s := sig.Sign(g, k, L, publicKeys, message, privateKeys)
	fmt.Printf("Signature (r=%s, s=%s)\n\n", r, s)
	fmt.Printf("Is the signature legit w.r.t the original public keys? %t\n\n", sig.Verify(g, L, message, r, s, publicKeys))

	privateKeys[1] = group.RandomScalar()
	publicKeys[1] = group.Mul(privateKeys[1], g)
	L = hashPublicKeys(publicKeys)

	fmt.Printf("Is the signature legit w.r.t a fake public key? %t\n", sig.Verify(g, L, message, r, s, publicKeys))
}
