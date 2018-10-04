package bn_test

import (
	"elrond-go-sandbox/crypto/bn"
	"elrond-go-sandbox/crypto/ed25519"
	"elrond-go-sandbox/crypto/math"
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

func TestSignVerify(t *testing.T) {

	signersCount := 3

	g := group.Generator()

	privateKeys := makeRandomScalars(signersCount)

	var publicKeys = make([]math.Point, len(privateKeys))

	for i := 0; i < len(privateKeys); i++ {
		publicKeys[i] = group.Mul(privateKeys[i], g)
	}

	message := "We're gonna be signing this!"

	k := makeRandomScalars(signersCount)

	L := sig.GetL(publicKeys)

	r, s := sig.Sign(g, k, L, publicKeys, message, privateKeys)

	isSignatureCorrect := sig.Verify(g, L, message, r, s, publicKeys)

	if !isSignatureCorrect {
		t.Errorf("Expected signature to be correct")
	}

	privateKeys[1] = group.RandomScalar()
	publicKeys[1] = group.Mul(privateKeys[1], g)
	L = sig.GetL(publicKeys)

	isSignatureCorrect = sig.Verify(g, L, message, r, s, publicKeys)

	if isSignatureCorrect {
		t.Errorf("Expected signature to be incorrect")
	}
}
