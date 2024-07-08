package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoncesTracker_computeExpectedSumOfNonces(t *testing.T) {
	tracker := newNoncesTracker()

	sum := tracker.computeExpectedSumOfNonces(0, 0)
	require.Equal(t, uint64(0), sum)

	sum = tracker.computeExpectedSumOfNonces(0, 1)
	require.Equal(t, uint64(0), sum)

	sum = tracker.computeExpectedSumOfNonces(0, 4)
	require.Equal(t, uint64(6), sum)

	sum = tracker.computeExpectedSumOfNonces(1, 4)
	require.Equal(t, uint64(10), sum)

	// https://www.wolframalpha.com/input?i=sum+of+consecutive+integers+between+100000+and+100041
	sum = tracker.computeExpectedSumOfNonces(100000, 42)
	require.Equal(t, uint64(4200861), sum)

	// https://www.wolframalpha.com/input?i=sum+of+consecutive+integers+between+1000000000000+and+1000000065534
	sum = tracker.computeExpectedSumOfNonces(oneTrillion, 65535)
	require.Equal(t, uint64(65535002147385345)%nonceModulus, sum)
}

func TestNoncesTracker_computeExpectedSumOfSquaresOfNonces(t *testing.T) {
	tracker := newNoncesTracker()

	sum := tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(0, 0)
	require.Equal(t, uint64(0), sum)

	sum = tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(0, 1)
	require.Equal(t, uint64(0), sum)

	sum = tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(0, 4)
	require.Equal(t, uint64(14)*six, sum)

	sum = tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(1, 4)
	require.Equal(t, uint64(30)*six, sum)

	// https://www.wolframalpha.com/input?i=sum+of+consecutive+squares+between+100000+and+100041
	sum = tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(100000, 42)
	require.Equal(t, (uint64(420172223821)*six)%nonceModulus, sum)

	// Python: (sum([i * i for i in range(1000000000, 1000065535)]) * 6) % 4294967295 = 92732025
	sum = tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(oneBillion, 65535)
	require.Equal(t, uint64(92732025), sum)

	// Python: (sum([i * i for i in range(1000000000000, 1000000000042)]) * 6) % 4294967295 = 307941426
	sum = tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(oneTrillion, 42)
	require.Equal(t, uint64(307941426), sum)

	// Python: (sum([i * i for i in range(1000000000000, 1000000065535)]) * 6) % 4294967295 = 445375860
	sum = tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(oneTrillion, 65535)
	require.Equal(t, uint64(445375860), sum)
}

func TestNoncesTracker_isSpotlessSequence(t *testing.T) {
	t.Run("empty sequence", func(t *testing.T) {
		tracker := newNoncesTracker()

		// A little bit of ambiguity (a sequence holding the nonce zero only behaves like an empty sequence):
		require.True(t, tracker.isSpotlessSequence(0, 0))
		require.True(t, tracker.isSpotlessSequence(0, 1))

		require.False(t, tracker.isSpotlessSequence(0, 2))
		require.False(t, tracker.isSpotlessSequence(7, 3))
	})

	t.Run("1-item sequence", func(t *testing.T) {
		tracker := newNoncesTracker()
		tracker.addNonce(0)

		// A little bit of ambiguity (a sequence holding the nonce zero only behaves like an empty sequence):
		require.True(t, tracker.isSpotlessSequence(0, 1))
		require.True(t, tracker.isSpotlessSequence(0, 0))

		require.False(t, tracker.isSpotlessSequence(0, 2))
		require.False(t, tracker.isSpotlessSequence(7, 3))

		tracker.removeNonce(0)
		tracker.addNonce(5)
		require.True(t, tracker.isSpotlessSequence(5, 1))
		require.False(t, tracker.isSpotlessSequence(5, 2))
		require.False(t, tracker.isSpotlessSequence(7, 1))
		require.False(t, tracker.isSpotlessSequence(7, 2))

		tracker.removeNonce(5)
		tracker.addNonce(42)
		require.True(t, tracker.isSpotlessSequence(42, 1))
		require.False(t, tracker.isSpotlessSequence(42, 2))
		require.False(t, tracker.isSpotlessSequence(7, 1))
		require.False(t, tracker.isSpotlessSequence(7, 2))
	})

	t.Run("with spotless addition and removal", func(t *testing.T) {
		t.Parallel()

		tracker := newNoncesTracker()
		numTotalTxsSender := uint64(100)
		firstNonce := uint64(oneBillion)
		lastNonce := firstNonce + numTotalTxsSender - 1
		numCurrentTxs := uint64(0)

		// We add nonces in increasing order:
		for nonce := firstNonce; nonce < firstNonce+numTotalTxsSender; nonce++ {
			tracker.addNonce(nonce)
			numCurrentTxs++

			isSpotless := tracker.isSpotlessSequence(firstNonce, numCurrentTxs)
			if !isSpotless {
				require.Fail(t, "nonce sequence is not spotless (after add)", "nonce: %d", nonce)
			}
		}

		// We remove nonces in decreasing order:
		for nonce := lastNonce; nonce >= firstNonce; nonce-- {
			tracker.removeNonce(nonce)
			numCurrentTxs--

			isSpotless := tracker.isSpotlessSequence(firstNonce, numCurrentTxs)
			if !isSpotless {
				require.Fail(t, "nonce sequence is not spotless (after remove)", "nonce: %d", nonce)
			}
		}
	})

	t.Run("with initial gap", func(t *testing.T) {
		tracker := newNoncesTracker()

		tracker.addNonce(5)
		tracker.addNonce(6)
		tracker.addNonce(7)

		require.False(t, tracker.isSpotlessSequence(2, 3))
	})

	t.Run("with initial duplicate", func(t *testing.T) {
		tracker := newNoncesTracker()

		tracker.addNonce(5)
		tracker.addNonce(5)
		tracker.addNonce(6)

		require.False(t, tracker.isSpotlessSequence(2, 3))
	})

	t.Run("with middle gap", func(t *testing.T) {
		tracker := newNoncesTracker()

		tracker.addNonce(5)
		tracker.addNonce(6)
		tracker.addNonce(8)

		require.False(t, tracker.isSpotlessSequence(5, 3))
	})

	t.Run("with middle duplicate", func(t *testing.T) {
		tracker := newNoncesTracker()

		tracker.addNonce(5)
		tracker.addNonce(6)
		tracker.addNonce(6)
		tracker.addNonce(8)

		require.False(t, tracker.isSpotlessSequence(5, 4))
		require.False(t, tracker.isSpotlessSequence(5, 3))
	})
}
