package executionTrack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonceHashes(t *testing.T) {
	t.Parallel()

	nh := newNonceHashes()

	nh.addNonceHash(1, "h1")
	nh.addNonceHash(1, "h1")
	nh.addNonceHash(1, "h2")

	require.Equal(t, 2, len(nh.getNonceHashes(1)))

	nh.removeFromNonceProvidedHash(1, "h1")
	require.Equal(t, []string{"h2"}, nh.getNonceHashes(1))
}

func TestGetNonceHashes_NonexistentNonce(t *testing.T) {
	nh := newNonceHashes()
	result := nh.getNonceHashes(42)
	require.Empty(t, result)
}

func TestRemoveFromNonceProvidedHash_NonexistentNonce(t *testing.T) {
	nh := newNonceHashes()
	// Should not panic or affect anything
	nh.addNonceHash(1, "h1")
	nh.removeFromNonceProvidedHash(99, "h1")
	require.Equal(t, []string{"h1"}, nh.getNonceHashes(1))
}

func TestAddRemove_MultipleNoncesIsolation(t *testing.T) {
	nh := newNonceHashes()

	nh.addNonceHash(1, "h1")
	nh.addNonceHash(2, "h2")
	nh.addNonceHash(2, "h3")

	nh.removeFromNonceProvidedHash(2, "h2")

	require.Equal(t, []string{"h1"}, nh.getNonceHashes(1))
	hashesNonce2 := nh.getNonceHashes(2)
	require.Len(t, hashesNonce2, 1)
	require.Equal(t, "h3", hashesNonce2[0])
}
