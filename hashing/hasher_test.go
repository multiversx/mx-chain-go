package hashing_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	blake2b "github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	sha256 "github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
)

func TestSha256(t *testing.T) {
	hashing.PutService(hashing.Hash, sha256.Sha256{})
	Suite(t, hashing.GetHasherService())
}

func TestBlake2b(t *testing.T) {
	hashing.PutService(hashing.Hash, blake2b.Blake2b{})
	Suite(t, hashing.GetHasherService())
}

func TestKeccak(t *testing.T) {
	hashing.PutService(hashing.Hash, keccak.Keccak{})
	Suite(t, hashing.GetHasherService())
}

func TestFnv(t *testing.T) {
	hashing.PutService(hashing.Hash, fnv.Fnv{})
	Suite(t, hashing.GetHasherService())
}

func Suite(t *testing.T, h hashing.Hasher) {
	TestingCalculateHash(t, h)
	TestingCalculateEmptyHash(t, h)
	TestingNilReturn(t, h)

}

func TestingCalculateHash(t *testing.T, h hashing.Hasher) {

	h1 := h.Compute("a")
	h2 := h.Compute("b")

	assert.NotEqual(t, h1, h2)

}

func TestingCalculateEmptyHash(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("")
	h2 := h.EmptyHash()

	assert.Equal(t, h1, h2)

}

func TestingNilReturn(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("a")
	assert.NotNil(t, h1)
}
