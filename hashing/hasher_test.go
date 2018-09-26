package hashing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	blake2b "github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	sha256 "github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
)

func TestSha256(t *testing.T) {
	hashing.PutService("Hasher", sha256.Sha256Impl{})
	Suite(t, hashing.GetHasherService())
}

func TestBlake2b(t *testing.T) {
	hashing.PutService("Hasher", blake2b.Blake2bImpl{})
	Suite(t, hashing.GetHasherService())
}

func Suite(t *testing.T, h hashing.Hasher) {
	TestingCalculateHash(t, h)
	TestingCalculateEmptyHash(t, h)
	TestingNilReturn(t, h)
	TestingCalculateHashWithNil(t, h)
}

func TestingCalculateHash(t *testing.T, h hashing.Hasher) {

	h1 := h.CalculateHash("a")
	h2 := h.CalculateHash("b")

	assert.NotEqual(t, h1, h2)

}

func TestingCalculateEmptyHash(t *testing.T, h hashing.Hasher) {
	h1 := h.CalculateHash("")
	h2 := h.EmptyHash()

	assert.Equal(t, h1, h2)

}

func TestingNilReturn(t *testing.T, h hashing.Hasher) {
	h1 := h.CalculateHash("a")
	assert.NotNil(t, h1)
}

func TestingCalculateHashWithNil(t *testing.T, h hashing.Hasher) {
	h1 := h.CalculateHash(nil)
	h2 := []byte{}

	assert.Equal(t, h1, h2)
}
