package schnorr_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/ed25519"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
)

var group = ed25519.Group{}
var hash = ed25519.Hash
var sig = schnorr.NewSig(group, hash)

func TestSignVerify(t *testing.T) {
	privateKey := group.RandomScalar()
	publicKey := group.Mul(privateKey, group.Generator())

	message := "We're gonna be signing this!"

	g := group.Generator()
	k := group.RandomScalar()

	r, s := sig.Sign(g, k, message, privateKey)

	derivedPublicKey := sig.PublicKey(g, message, r, s)

	if !group.Equal(publicKey, derivedPublicKey) {
		t.Errorf("Expected public key %d and derived public key %d to be equal", publicKey, derivedPublicKey)
	}

	isSignatureCorrect := sig.Verify(g, message, r, s, publicKey)

	if !isSignatureCorrect {
		t.Errorf("Expected signature to be correct")
	}

	fakePrivateKey := group.RandomScalar()
	fakePublicKey := group.Mul(fakePrivateKey, group.Generator())

	isSignatureCorrect = sig.Verify(g, message, r, s, fakePublicKey)

	if isSignatureCorrect {
		t.Errorf("Expected signature to be incorrect")
	}
}
