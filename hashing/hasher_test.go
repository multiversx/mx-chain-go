package hashing_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

func TestSha256(t *testing.T) {
	Suite(t, &hashing.Sha256{})
}

func TestBlake2b(t *testing.T) {
	Suite(t, &hashing.Blake2b{})
}

func Suite(t *testing.T, h hashing.Hasher) {
	TestingHasher(t, h)
}

func TestingHasher(t *testing.T, h hashing.Hasher) {

	h1 := h.CalculateHash("a")
	h2 := h.CalculateHash("b")

	if h1 == h2 {
		t.Fatal("Same hash for different arguments")
	}

	fmt.Printf("Hash 1 = %s\nHash 2 = %s\n", h1, h2)
}
