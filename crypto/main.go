package main

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

import (
	"elrond-go-sandbox/crypto/ed25519"
	"elrond-go-sandbox/crypto/schnorr"
	"fmt"
)

var group = ed25519.Group{}

func main() {
	privateKey := group.RandomScalar()
	publicKey := group.Mul(privateKey, group.Generator())

	fmt.Printf("Generated private key: %s\n", privateKey)
	fmt.Printf("Derived public key: %s\n\n", publicKey)

	message := "We're gonna be signing this!"

	r, s := schnorr.Sign(group, message, privateKey, ed25519.Hash)
	fmt.Printf("Signature (r=%s, s=%s)\n\n", r, s)

	derivedPublicKey := schnorr.PublicKey(group, message, r, s, ed25519.Hash)
	fmt.Printf("Derived public key: %s\n", derivedPublicKey)
	fmt.Printf("Are the original and derived public keys the same? %t\n", publicKey.Equal(derivedPublicKey))
	fmt.Printf("Is the signature legit w.r.t the original public key? %t\n\n", schnorr.Verify(group, message, r, s, publicKey, ed25519.Hash))

	fakePrivateKey := group.RandomScalar()
	fakePublicKey := group.Mul(fakePrivateKey, group.Generator())
	fmt.Printf("Fake public key: %s\n", fakePublicKey)
	fmt.Printf("Is the signature legit w.r.t a fake public key? %t\n", schnorr.Verify(group, message, r, s, fakePublicKey, ed25519.Hash))
}
