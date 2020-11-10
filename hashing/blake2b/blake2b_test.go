package blake2b_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	blake2bLib "golang.org/x/crypto/blake2b"
)

func TestNewBlake2bWithSizeInvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	h, err := blake2b.NewBlake2bWithSize(-2)
	require.Nil(t, h)
	require.Equal(t, blake2b.ErrInvalidHashSize, err)
}

func TestBlake2b_ComputeWithDifferentHashSizes(t *testing.T) {
	t.Parallel()

	input := "dummy string"
	sizes := []int{2, 5, 8, 16, 32, 37, 64}
	for _, size := range sizes {
		testComputeOk(t, input, size)
	}
}

func testComputeOk(t *testing.T, input string, size int) {
	hasher, err := blake2b.NewBlake2bWithSize(size)
	require.NoError(t, err)
	res := hasher.Compute(input)
	assert.Equal(t, size, len(res))
}

func TestBlake2b_Empty(t *testing.T) {

	hasher, err := blake2b.NewBlake2bWithSize(64)
	require.NoError(t, err)

	var nilStr string
	nilRes := hasher.Compute(nilStr)
	assert.Equal(t, 64, len(nilRes))

	emptyRes := hasher.Compute("")
	assert.Equal(t, 64, len(emptyRes))

	assert.Equal(t, emptyRes, nilRes)

	// force recompute
	hasher, _ = blake2b.NewBlake2bWithSize(64)

	emptyRes = hasher.Compute("")
	assert.Equal(t, 64, len(emptyRes))

	assert.Equal(t, emptyRes, nilRes)
}

func TestBlake2b_Size(t *testing.T) {
	t.Parallel()

	h1 := blake2b.NewBlake2b()
	require.Equal(t, blake2bLib.Size256, h1.Size())

	customSize := 37
	h2, err := blake2b.NewBlake2bWithSize(customSize)
	require.NoError(t, err)
	require.Equal(t, customSize, h2.Size())
}
