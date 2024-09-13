package proofscache_test

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProofsPool(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool()
	require.False(t, pp.IsInterfaceNil())
}

func TestProofsPool_ShouldWork(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)

	pp := proofscache.NewProofsPool()

	proof1 := &block.HeaderProof{
		PubKeysBitmap:       []byte("pubKeysBitmap1"),
		AggregatedSignature: []byte("aggSig1"),
		HeaderHash:          []byte("hash1"),
		HeaderEpoch:         1,
		HeaderNonce:         1,
		HeaderShardId:       shardID,
	}
	proof2 := &block.HeaderProof{
		PubKeysBitmap:       []byte("pubKeysBitmap2"),
		AggregatedSignature: []byte("aggSig2"),
		HeaderHash:          []byte("hash2"),
		HeaderEpoch:         1,
		HeaderNonce:         2,
		HeaderShardId:       shardID,
	}
	proof3 := &block.HeaderProof{
		PubKeysBitmap:       []byte("pubKeysBitmap3"),
		AggregatedSignature: []byte("aggSig3"),
		HeaderHash:          []byte("hash3"),
		HeaderEpoch:         1,
		HeaderNonce:         3,
		HeaderShardId:       shardID,
	}
	proof4 := &block.HeaderProof{
		PubKeysBitmap:       []byte("pubKeysBitmap4"),
		AggregatedSignature: []byte("aggSig4"),
		HeaderHash:          []byte("hash4"),
		HeaderEpoch:         1,
		HeaderNonce:         4,
		HeaderShardId:       shardID,
	}
	_ = pp.AddProof(proof1)
	_ = pp.AddProof(proof2)
	_ = pp.AddProof(proof3)
	_ = pp.AddProof(proof4)

	proof, err := pp.GetProof(shardID, []byte("hash3"))
	require.Nil(t, err)
	require.Equal(t, proof3, proof)

	err = pp.CleanupProofsBehindNonce(shardID, 4)
	require.Nil(t, err)

	proof, err = pp.GetProof(shardID, []byte("hash3"))
	require.Equal(t, proofscache.ErrMissingProof, err)
	require.Nil(t, proof)

	proof, err = pp.GetProof(shardID, []byte("hash4"))
	require.Nil(t, err)
	require.Equal(t, proof4, proof)
}

func TestProofsPool_Concurrency(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool()

	numOperations := 1000

	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	cnt := uint32(0)

	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			switch idx % 5 {
			case 0, 1, 2:
				_ = pp.AddProof(generateProof())
			case 3:
				_, err := pp.GetProof(generateRandomShardID(), generateRandomHash())
				if errors.Is(err, proofscache.ErrMissingProof) {
					atomic.AddUint32(&cnt, 1)
				}
			case 4:
				_ = pp.CleanupProofsBehindNonce(generateRandomShardID(), generateRandomNonce())
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
		HeaderNonce:   generateRandomNonce(),
		HeaderShardId: generateRandomShardID(),
	}
}

func generateRandomHash() []byte {
	hashSuffix := generateRandomInt(100)
	hash := []byte("hash_" + hashSuffix.String())
	return hash
}

func generateRandomNonce() uint64 {
	val := generateRandomInt(3)
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
