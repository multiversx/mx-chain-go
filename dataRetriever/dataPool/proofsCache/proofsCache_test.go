package proofscache_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProofsCache(t *testing.T) {
	t.Parallel()

	t.Run("incremental nonces, should cleanup all caches", func(t *testing.T) {
		t.Parallel()

		proof0 := &block.HeaderProof{HeaderHash: []byte{0}, HeaderNonce: 0}
		proof1 := &block.HeaderProof{HeaderHash: []byte{1}, HeaderNonce: 1}
		proof2 := &block.HeaderProof{HeaderHash: []byte{2}, HeaderNonce: 2}
		proof3 := &block.HeaderProof{HeaderHash: []byte{3}, HeaderNonce: 3}
		proof4 := &block.HeaderProof{HeaderHash: []byte{4}, HeaderNonce: 4}

		pc := proofscache.NewProofsCache(4)

		pc.AddProof(proof0)
		pc.AddProof(proof1)
		pc.AddProof(proof2)
		pc.AddProof(proof3)

		require.Equal(t, 4, pc.FullProofsByNonceSize())
		require.Equal(t, 4, pc.ProofsByHashSize())

		pc.AddProof(proof4) // added to new head bucket

		require.Equal(t, 5, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(4)
		require.Equal(t, 1, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(10)
		require.Equal(t, 0, pc.ProofsByHashSize())
	})

	t.Run("non incremental nonces", func(t *testing.T) {
		t.Parallel()

		proof0 := &block.HeaderProof{HeaderHash: []byte{0}, HeaderNonce: 0}
		proof1 := &block.HeaderProof{HeaderHash: []byte{1}, HeaderNonce: 1}
		proof2 := &block.HeaderProof{HeaderHash: []byte{2}, HeaderNonce: 2}
		proof3 := &block.HeaderProof{HeaderHash: []byte{3}, HeaderNonce: 3}
		proof4 := &block.HeaderProof{HeaderHash: []byte{4}, HeaderNonce: 4}
		proof5 := &block.HeaderProof{HeaderHash: []byte{5}, HeaderNonce: 5}

		pc := proofscache.NewProofsCache(4)

		pc.AddProof(proof4)
		pc.AddProof(proof1)
		pc.AddProof(proof2)
		pc.AddProof(proof3)

		require.Equal(t, 4, pc.FullProofsByNonceSize())
		require.Equal(t, 4, pc.ProofsByHashSize())

		pc.AddProof(proof0) // added to new head bucket

		require.Equal(t, 5, pc.FullProofsByNonceSize())
		require.Equal(t, 5, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(4)

		// cleanup up head bucket with only one proof
		require.Equal(t, 1, pc.ProofsByHashSize())

		pc.AddProof(proof5) // added to new head bucket

		require.Equal(t, 2, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(5) // will not remove any bucket
		require.Equal(t, 2, pc.FullProofsByNonceSize())
		require.Equal(t, 2, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(10)
		require.Equal(t, 0, pc.ProofsByHashSize())
	})

	t.Run("shuffled nonces, should cleanup all caches", func(t *testing.T) {
		t.Parallel()

		pc := proofscache.NewProofsCache(10)

		nonces := generateShuffledNonces(100)
		for _, nonce := range nonces {
			proof := generateProof()
			proof.HeaderNonce = nonce
			proof.HeaderHash = []byte("hash_" + fmt.Sprintf("%d", nonce))

			pc.AddProof(proof)
		}

		require.Equal(t, 100, pc.FullProofsByNonceSize())
		require.Equal(t, 100, pc.ProofsByHashSize())

		pc.CleanupProofsBehindNonce(100)
		require.Equal(t, 0, pc.FullProofsByNonceSize())
		require.Equal(t, 0, pc.ProofsByHashSize())
	})
}

func TestProofsCache_Concurrency(t *testing.T) {
	t.Parallel()

	pc := proofscache.NewProofsCache(100)

	numOperations := 1000

	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			switch idx % 2 {
			case 0:
				pc.AddProof(generateProof())
			case 1:
				pc.CleanupProofsBehindNonce(generateRandomNonce(100))
			default:
				assert.Fail(t, "should have not beed called")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
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
