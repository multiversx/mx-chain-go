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
