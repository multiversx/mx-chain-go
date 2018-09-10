package main

// https://medium.com/coinmonks/schnorr-signatures-in-go-80a7fbfe0fe4

import (
	"elrond-go-sandbox/crypto/schnorr"
	"fmt"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
)

var curve = edwards25519.NewBlakeSHA256Ed25519()

func main() {
	privateKey := curve.Scalar().Pick(curve.RandomStream())
	publicKey := curve.Point().Mul(privateKey, curve.Point().Base())

	fmt.Printf("Generated private key: %s\n", privateKey)
	fmt.Printf("Derived public key: %s\n\n", publicKey)

	message := "We're gonna be signing this!"

	signature := schnorr.Sign(message, privateKey)
	fmt.Printf("Signature %s\n\n", signature)

	derivedPublicKey := schnorr.PublicKey(message, signature)
	fmt.Printf("Derived public key: %s\n", derivedPublicKey)
	fmt.Printf("Are the original and derived public keys the same? %t\n", publicKey.Equal(derivedPublicKey))
	fmt.Printf("Is the signature legit w.r.t the original public key? %t\n\n", schnorr.Verify(message, signature, publicKey))

	fakePublicKey := curve.Point().Mul(curve.Scalar().Neg(curve.Scalar().One()), publicKey)
	fmt.Printf("Fake public key: %s\n", fakePublicKey)
	fmt.Printf("Is the signature legit w.r.t a fake public key? %t\n", schnorr.Verify(message, signature, fakePublicKey))
}
