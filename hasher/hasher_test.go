package hasher

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/hasher/sha256"
	"testing"
)

func TestHasher(t *testing.T) {

	si := hasher.Sha256Impl{}

	h1 := si.CalculateHash("a")
	h2 := si.CalculateHash("b")

	if h1 == h2 {
		t.Fatal("Same hash for different arguments")
	}

	fmt.Printf("Hash 1 = %s\nHash 2 = %s\n", h1, h2)
}
