package main

import (
	"fmt"
	"gopkg.in/dedis/kyber.v2/group/edwards25519"
)

var curve = edwards25519.NewBlakeSHA256Ed25519()

func main() {
	privateKey := curve.Scalar().Pick(curve.RandomStream())
	publicKey := curve.Point().Mul(privateKey, curve.Point().Base())

	fmt.Printf("Generated private key: %s\n", privateKey)
	fmt.Printf("Derived public key: %s\n\n", publicKey)
}
