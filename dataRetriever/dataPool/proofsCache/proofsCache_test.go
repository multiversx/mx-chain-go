package proofscache_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/stretchr/testify/require"
)

func TestProofsCache(t *testing.T) {
	t.Parallel()

	t.Run("incremental nonces, should cleanup all caches", func(t *testing.T) {
		t.Parallel()

		proof1 := &block.HeaderProof{HeaderHash: []byte{1}, HeaderNonce: 1}
		proof2 := &block.HeaderProof{HeaderHash: []byte{2}, HeaderNonce: 2}
		proof3 := &block.HeaderProof{HeaderHash: []byte{3}, HeaderNonce: 3}
		proof4 := &block.HeaderProof{HeaderHash: []byte{4}, HeaderNonce: 4}

		pc := proofscache.NewProofsCache()

		pc.AddProof(proof1)
		pc.AddProof(proof2)
		pc.AddProof(proof3)
		pc.AddProof(proof4)

		require.Equal(t, 4, pc.ProofsByNonceSize())
		require.Equal(t, 4, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(4)
		require.Equal(t, 1, pc.ProofsByNonceSize())
		require.Equal(t, 1, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(10)
		require.Equal(t, 0, pc.ProofsByNonceSize())
		require.Equal(t, 0, pc.ProofsByHashSize())
	})

	t.Run("shuffled nonces, should cleanup all caches", func(t *testing.T) {
		t.Parallel()

		pc := proofscache.NewProofsCache()

		nonces := generateShuffledNonces(100)
		for _, nonce := range nonces {
			proof := generateProof()
			proof.HeaderNonce = nonce
			proof.HeaderHash = []byte("hash_" + fmt.Sprintf("%d", nonce))

			pc.AddProof(proof)
		}

		require.Equal(t, 100, pc.ProofsByNonceSize())
		require.Equal(t, 100, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(50)
		require.Equal(t, 50, pc.ProofsByNonceSize())
		require.Equal(t, 50, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(100)
		require.Equal(t, 0, pc.ProofsByNonceSize())
		require.Equal(t, 0, pc.ProofsByHashSize())
	})
}

func generateShuffledNonces(n int) []uint64 {
	nonces := make([]uint64, n)
	for i := uint64(0); i < uint64(n); i++ {
		nonces[i] = i
	}

	rand.Shuffle(len(nonces), func(i, j int) {
		nonces[i], nonces[j] = nonces[j], nonces[i]
	})

	return nonces
}
