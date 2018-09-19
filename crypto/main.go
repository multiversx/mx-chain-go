package main

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

import (
	"elrond-go-sandbox/crypto/ed25519"
	"elrond-go-sandbox/crypto/schnorr"
	"fmt"
)

var group = ed25519.Group{}
var hash = ed25519.Hash
var sig = schnorr.NewSig(group, hash)

func main() {
	privateKey := group.RandomScalar()
	publicKey := group.Mul(privateKey, group.Generator())

	fmt.Printf("Generated private key: %s\n", privateKey)
	fmt.Printf("Derived public key: %s\n\n", publicKey)

	message := "We're gonna be signing this!"

	g := group.Generator()
	k := group.RandomScalar()

	r, s := sig.Sign(g, k, message, privateKey)
	fmt.Printf("Signature (r=%s, s=%s)\n\n", r, s)

	derivedPublicKey := sig.PublicKey(g, message, r, s)
	fmt.Printf("Derived public key: %s\n", derivedPublicKey)
	fmt.Printf("Are the original and derived public keys the same? %t\n", group.Equal(publicKey, derivedPublicKey))
	fmt.Printf("Is the signature legit w.r.t the original public key? %t\n\n", sig.Verify(g, message, r, s, publicKey))

	fakePrivateKey := group.RandomScalar()
	fakePublicKey := group.Mul(fakePrivateKey, group.Generator())
	fmt.Printf("Fake public key: %s\n", fakePublicKey)
	fmt.Printf("Is the signature legit w.r.t a fake public key? %t\n", sig.Verify(g, message, r, s, fakePublicKey))
}
