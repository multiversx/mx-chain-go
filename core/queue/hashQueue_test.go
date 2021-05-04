package queue

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHashQueue(t *testing.T) {
	t.Parallel()

	hqSize := uint(5)
	hq := NewSliceQueue(hqSize)
	assert.Equal(t, hqSize, hq.size)
	assert.Equal(t, 0, len(hq.queue))
}

func TestSliceQueue_AddToAnZeroSizedQueue(t *testing.T) {
	t.Parallel()

	hashToAdd := []byte("hash")
	hq := NewSliceQueue(0)

	returnedHash := hq.Add(hashToAdd)
	assert.Equal(t, hashToAdd, returnedHash)
	assert.Equal(t, 0, len(hq.queue))
}

func TestSliceQueue_AddToAFullQueue(t *testing.T) {
	t.Parallel()

	hashToAdd := []byte("hash")
	expectedHash := []byte("expectedHash")
	hqSize := uint(5)
	hq := NewSliceQueue(hqSize)

	returnedHash := hq.Add(expectedHash)
	assert.Equal(t, 0, len(returnedHash))

	for i := 0; i < int(hqSize-1); i++ {
		returnedHash = hq.Add([]byte(strconv.Itoa(i)))
		assert.Equal(t, 0, len(returnedHash))
	}

	assert.Equal(t, int(hq.size), len(hq.queue))

	returnedHash = hq.Add(hashToAdd)
	assert.Equal(t, expectedHash, returnedHash)
	assert.Equal(t, hashToAdd, hq.queue[len(hq.queue)-1])
}

func TestSliceQueue_AddToPartiallyFilledQueue(t *testing.T) {
	t.Parallel()

	hashToAdd := []byte("hash")
	hqSize := uint(5)
	hq := NewSliceQueue(hqSize)

	numHashesToAdd := 3
	for i := 0; i < numHashesToAdd; i++ {
		returnedHash := hq.Add([]byte(strconv.Itoa(i)))
		assert.Equal(t, 0, len(returnedHash))
	}

	assert.Equal(t, numHashesToAdd, len(hq.queue))

	returnedHash := hq.Add(hashToAdd)
	assert.Equal(t, 0, len(returnedHash))
	assert.Equal(t, numHashesToAdd+1, len(hq.queue))
}
