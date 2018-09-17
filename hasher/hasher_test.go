package hasher

import (
	"fmt"
	"testing"
)

func TestHasher(t *testing.T) {

	var sha256 Sha256Impl

	h1 := sha256.CalculateHash("a")
	h2 := sha256.CalculateHash("b")

	if h1 == h2 {
		t.Fatal("Same hash for different arguments")
	}

	fmt.Printf("Hash 1 = %s\nHash 2 = %s\n", h1, h2)
}
