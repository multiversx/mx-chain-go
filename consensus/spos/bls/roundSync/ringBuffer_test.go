package roundSync

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundRingBuffer_AddAndLen(t *testing.T) {
	t.Parallel()

	rb := newRoundRingBuffer(3)

	rb.add(10)
	rb.add(20)
	rb.add(30)
	require.Equal(t, 3, rb.len())

	rb.add(20)
	require.Equal(t, 3, rb.len())
	require.Equal(t, []uint64{10, 20, 30}, rb.last(5))
}

func TestRoundRingBuffer_OverwriteOldest(t *testing.T) {
	t.Parallel()

	rb := newRoundRingBuffer(2)

	rb.add(1)
	rb.add(2)
	require.Equal(t, []uint64{1, 2}, rb.last(2))

	// This should overwrite "1"
	rb.add(3)
	require.Equal(t, 2, rb.len())
	require.Equal(t, []uint64{2, 3}, rb.last(2))

	// "1" must no longer exist
	require.False(t, rb.contains(1))
	require.True(t, rb.contains(2))
	require.True(t, rb.contains(3))
}

func TestRoundRingBuffer_Last(t *testing.T) {
	t.Parallel()

	rb := newRoundRingBuffer(5)

	rb.add(5)
	rb.add(6)
	rb.add(7)

	last := rb.last(2)
	require.Equal(t, []uint64{6, 7}, last)

	// Requesting more than size should return everything
	last = rb.last(10)
	require.Equal(t, []uint64{5, 6, 7}, last)
}
