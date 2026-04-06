package executionTrack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNonceHash(t *testing.T) {
	t.Parallel()

	nh := newNonceHash()

	nh.addNonceHash(1, "h1")
	nh.addNonceHash(2, "h2")
	nh.addNonceHash(3, "h3")

	nh.removeByNonce(2)

	require.Equal(t, "h1", nh.getHashByNonce(1))
	require.Equal(t, "", nh.getHashByNonce(2))
	require.Equal(t, "h3", nh.getHashByNonce(3))
}
