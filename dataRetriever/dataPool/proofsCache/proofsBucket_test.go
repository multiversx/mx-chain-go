package proofscache_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
)

func TestProofNonceBucket_Remove(t *testing.T) {
	t.Parallel()

	t.Run("removing one of several hashes should keep the nonce", func(t *testing.T) {
		t.Parallel()

		bucket := proofscache.NewProofBucket()
		bucket.Insert(5, "hash1")
		bucket.Insert(5, "hash2")

		bucket.Remove(5, "hash1")

		require.Equal(t, []string{"hash2"}, bucket.HashesAt(5))
		require.Equal(t, 1, bucket.TrackedNonces())
		require.Equal(t, uint64(5), bucket.MaxNonce())
	})

	t.Run("removing the last hash should drop the nonce", func(t *testing.T) {
		t.Parallel()

		bucket := proofscache.NewProofBucket()
		bucket.Insert(5, "hash1")
		bucket.Insert(7, "hash2")

		bucket.Remove(5, "hash1")

		require.Empty(t, bucket.HashesAt(5))
		require.Equal(t, 1, bucket.TrackedNonces())
		require.Equal(t, 1, bucket.Size())
	})

	t.Run("emptying the highest nonce should lower maxNonce", func(t *testing.T) {
		t.Parallel()

		bucket := proofscache.NewProofBucket()
		bucket.Insert(5, "hash1")
		bucket.Insert(7, "hash2")
		require.Equal(t, uint64(7), bucket.MaxNonce())

		bucket.Remove(7, "hash2")

		require.Equal(t, uint64(5), bucket.MaxNonce())
		require.Equal(t, 1, bucket.TrackedNonces())
	})

	t.Run("emptying a lower nonce should keep maxNonce", func(t *testing.T) {
		t.Parallel()

		bucket := proofscache.NewProofBucket()
		bucket.Insert(5, "hash1")
		bucket.Insert(7, "hash2")

		bucket.Remove(5, "hash1")

		require.Equal(t, uint64(7), bucket.MaxNonce())
	})

	t.Run("emptying the bucket should reset maxNonce", func(t *testing.T) {
		t.Parallel()

		bucket := proofscache.NewProofBucket()
		bucket.Insert(5, "hash1")

		bucket.Remove(5, "hash1")

		require.Equal(t, uint64(0), bucket.MaxNonce())
		require.Equal(t, 0, bucket.TrackedNonces())
		require.Equal(t, 0, bucket.Size())
	})

	t.Run("reinserting after the bucket emptied should raise maxNonce again", func(t *testing.T) {
		t.Parallel()

		bucket := proofscache.NewProofBucket()
		bucket.Insert(5, "hash1")
		bucket.Remove(5, "hash1")

		bucket.Insert(9, "hash2")

		require.Equal(t, uint64(9), bucket.MaxNonce())
		require.Equal(t, 1, bucket.Size())
	})

	t.Run("removing a missing hash should be a no-op", func(t *testing.T) {
		t.Parallel()

		bucket := proofscache.NewProofBucket()
		bucket.Insert(5, "hash1")

		bucket.Remove(5, "unknown")
		bucket.Remove(9, "hash1")

		require.Equal(t, []string{"hash1"}, bucket.HashesAt(5))
		require.Equal(t, uint64(5), bucket.MaxNonce())
		require.Equal(t, 1, bucket.TrackedNonces())
	})
}
