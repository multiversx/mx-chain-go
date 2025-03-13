package proofscache_test

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cleanupDelta = 3

var shardID = uint32(1)

var proof1 = &block.HeaderProof{
	PubKeysBitmap:       []byte("pubKeysBitmap1"),
	AggregatedSignature: []byte("aggSig1"),
	HeaderHash:          []byte("hash1"),
	HeaderEpoch:         1,
	HeaderNonce:         1,
	HeaderShardId:       shardID,
}

var proof2 = &block.HeaderProof{
	PubKeysBitmap:       []byte("pubKeysBitmap2"),
	AggregatedSignature: []byte("aggSig2"),
	HeaderHash:          []byte("hash2"),
	HeaderEpoch:         1,
	HeaderNonce:         2,
	HeaderShardId:       shardID,
}
var proof3 = &block.HeaderProof{
	PubKeysBitmap:       []byte("pubKeysBitmap3"),
	AggregatedSignature: []byte("aggSig3"),
	HeaderHash:          []byte("hash3"),
	HeaderEpoch:         1,
	HeaderNonce:         3,
	HeaderShardId:       shardID,
}
var proof4 = &block.HeaderProof{
	PubKeysBitmap:       []byte("pubKeysBitmap4"),
	AggregatedSignature: []byte("aggSig4"),
	HeaderHash:          []byte("hash4"),
	HeaderEpoch:         1,
	HeaderNonce:         4,
	HeaderShardId:       shardID,
}

func TestNewProofsPool(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta)
	require.False(t, pp.IsInterfaceNil())
}

func TestProofsPool_ShouldWork(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta)

	_ = pp.AddProof(proof1)
	_ = pp.AddProof(proof2)
	_ = pp.AddProof(proof3)
	_ = pp.AddProof(proof4)

	ok := pp.AddProof(proof4)
	require.False(t, ok)

	proof, err := pp.GetProof(shardID, []byte("hash3"))
	require.Nil(t, err)
	require.Equal(t, proof3, proof)

	err = pp.CleanupProofsBehindNonce(shardID, 4)
	require.Nil(t, err)

	proof, err = pp.GetProof(shardID, []byte("hash3"))
	require.Nil(t, err)
	require.Equal(t, proof3, proof)

	proof, err = pp.GetProof(shardID, []byte("hash4"))
	require.Nil(t, err)
	require.Equal(t, proof4, proof)
}

func TestProofsPool_RegisterHandler(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta)

	wasCalled := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	handler := func(proof data.HeaderProofHandler) {
		wasCalled = true
		wg.Done()
	}
	pp.RegisterHandler(nil)
	pp.RegisterHandler(handler)

	_ = pp.AddProof(generateProof())

	wg.Wait()

	assert.True(t, wasCalled)
}

func TestProofsPool_CleanupProofsBehindNonce(t *testing.T) {
	t.Parallel()

	t.Run("should not cleanup proofs behind delta", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta)

		_ = pp.AddProof(proof1)
		_ = pp.AddProof(proof2)
		_ = pp.AddProof(proof3)
		_ = pp.AddProof(proof4)

		err := pp.CleanupProofsBehindNonce(shardID, 5)
		require.Nil(t, err)

		proof, err := pp.GetProof(shardID, []byte("hash1"))
		require.Equal(t, proofscache.ErrMissingProof, err)
		require.Nil(t, proof)

		_, err = pp.GetProof(shardID, []byte("hash2"))
		require.Nil(t, err)
		_, err = pp.GetProof(shardID, []byte("hash3"))
		require.Nil(t, err)
		_, err = pp.GetProof(shardID, []byte("hash4"))
		require.Nil(t, err)
	})

	t.Run("should not cleanup if nonce smaller or equal to delta", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta)

		_ = pp.AddProof(proof1)
		_ = pp.AddProof(proof2)
		_ = pp.AddProof(proof3)
		_ = pp.AddProof(proof4)

		err := pp.CleanupProofsBehindNonce(shardID, cleanupDelta)
		require.Nil(t, err)

		_, err = pp.GetProof(shardID, []byte("hash1"))
		require.Nil(t, err)
		_, err = pp.GetProof(shardID, []byte("hash2"))
		require.Nil(t, err)
		_, err = pp.GetProof(shardID, []byte("hash3"))
		require.Nil(t, err)
		_, err = pp.GetProof(shardID, []byte("hash4"))
		require.Nil(t, err)
	})
}

func TestProofsPool_Concurrency(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta)

	numOperations := 1000

	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	cnt := uint32(0)

	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			switch idx % 6 {
			case 0, 1, 2:
				_ = pp.AddProof(generateProof())
			case 3:
				_, err := pp.GetProof(generateRandomShardID(), generateRandomHash())
				if errors.Is(err, proofscache.ErrMissingProof) {
					atomic.AddUint32(&cnt, 1)
				}
			case 4:
				_ = pp.CleanupProofsBehindNonce(generateRandomShardID(), generateRandomNonce(100))
			case 5:
				handler := func(proof data.HeaderProofHandler) {
				}
				pp.RegisterHandler(handler)
			default:
				assert.Fail(t, "should have not beed called")
			}

			wg.Done()
		}(i)
	}

	require.GreaterOrEqual(t, uint32(numOperations/3), atomic.LoadUint32(&cnt))
}

func generateProof() *block.HeaderProof {
	return &block.HeaderProof{
		HeaderHash:    generateRandomHash(),
		HeaderEpoch:   1,
		HeaderNonce:   generateRandomNonce(100),
		HeaderShardId: generateRandomShardID(),
	}
}

func generateRandomHash() []byte {
	hashSuffix := generateRandomInt(100)
	hash := []byte("hash_" + hashSuffix.String())
	return hash
}

func generateRandomNonce(n int64) uint64 {
	val := generateRandomInt(n)
	return val.Uint64()
}

func generateRandomShardID() uint32 {
	val := generateRandomInt(3)
	return uint32(val.Uint64())
}

func generateRandomInt(max int64) *big.Int {
	rantInt, _ := rand.Int(rand.Reader, big.NewInt(max))
	return rantInt
}
