package hasher

import (
	"fmt"
	"testing"
)

func TestHasher(t *testing.T) {

	hs := GetHasherService()

	h1 := hs.CalculateHash("a")
	h2 := hs.CalculateHash("b")

	if h1 == h2 {
		t.Fatal("Same hash for different arguments")
	}

	fmt.Printf("Hash 1 = %s\nHash 2 = %s\n", h1, h2)
}
