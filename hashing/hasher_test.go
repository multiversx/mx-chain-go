package hashing_test

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSha256(t *testing.T) {
	Suite(t, sha256.NewSha256())
}

func TestBlake2b(t *testing.T) {
	Suite(t, blake2b.NewBlake2b())
}

func TestKeccak(t *testing.T) {
	Suite(t, keccak.NewKeccak())
}

func TestFnv(t *testing.T) {
	Suite(t, fnv.NewFnv())
}

func Suite(t *testing.T, h hashing.Hasher) {
	testNilInterface(t, h)
	testSize(t, h)
	testCalculateHash(t, h)
	testCalculateEmptyHash(t, h)
	testNilReturn(t, h)
	testConcurrentSafe(t, h)
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

	assert.Equal(t, len(h1), h.Size())
}

func testNilReturn(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("a")

	assert.NotNil(t, h1)
}

func testConcurrentSafe(t *testing.T, h hashing.Hasher) {
	numGoRoutines := 10
	wg := sync.WaitGroup{}
	wg.Add(numGoRoutines)

	for i := 0; i < numGoRoutines; i++ {
		go func(wg *sync.WaitGroup) {
			defer func() {
				r := recover()
				require.Nil(t, r)
			}()

			h.Compute("a")
			h.Size()
			wg.Done()
		}(&wg)
	}

	wg.Wait()
}
