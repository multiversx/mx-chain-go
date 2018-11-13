package hashing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

func TestSha256(t *testing.T) {
	Suite(t, hashing.Sha256{})
}

func TestBlake2b(t *testing.T) {
	Suite(t, hashing.Blake2b{})
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
