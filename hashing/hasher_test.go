package hashing_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/stretchr/testify/assert"
)

func TestSha256(t *testing.T) {
	Suite(t, sha256.Sha256{})
}

func TestBlake2b(t *testing.T) {
	Suite(t, &blake2b.Blake2b{})
}

func TestKeccak(t *testing.T) {
	Suite(t, keccak.Keccak{})
}

func TestFnv(t *testing.T) {
	Suite(t, fnv.Fnv{})
}

func Suite(t *testing.T, h hashing.Hasher) {
	testNilInterface(t, h)
	testSize(t, h)
	testCalculateHash(t, h)
	testCalculateEmptyHash(t, h)
	testNilReturn(t, h)
}

func testNilInterface(t *testing.T, h hashing.Hasher) {
	res := h.IsInterfaceNil()

	assert.False(t, res)
}

func testSize(t *testing.T, h hashing.Hasher) {
	input := "test"
	res := h.Compute(input)
	hasherSize := h.Size()

	assert.Equal(t, hasherSize, len(res))
}

func testCalculateHash(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("a")
	h2 := h.Compute("b")

	assert.NotEqual(t, h1, h2)
}

func testCalculateEmptyHash(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("")
	h2 := h.EmptyHash()

	assert.Equal(t, h1, h2)
	assert.Equal(t, h.Size(), len(h1))
}

func testNilReturn(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("a")
	assert.NotNil(t, h1)
}
