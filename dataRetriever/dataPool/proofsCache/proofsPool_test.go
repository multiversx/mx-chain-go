package proofscache_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
)

const cleanupDelta = 3
const bucketSize = 100

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

	pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)
	require.False(t, pp.IsInterfaceNil())
}

func TestProofsPool_ShouldWork(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

	ok := pp.AddProof(nil)
	require.False(t, ok)

	_ = pp.AddProof(proof1)
	_ = pp.AddProof(proof2)
	_ = pp.AddProof(proof3)
	_ = pp.AddProof(proof4)

	ok = pp.AddProof(proof4)
	require.False(t, ok)

	proof, err := pp.GetProof(shardID, []byte("hash3"))
	require.Nil(t, err)
	require.Equal(t, proof3, proof)
	proof, err = pp.GetProofByNonce(3, shardID)
	require.Nil(t, err)
	require.Equal(t, proof3, proof)

	err = pp.CleanupProofsBehindNonce(shardID, 4)
	require.Nil(t, err)

	proof, err = pp.GetProof(shardID, []byte("hash3"))
	require.Nil(t, err)
	require.Equal(t, proof3, proof)
	proof, err = pp.GetProofByNonce(3, shardID)
	require.Nil(t, err)
	require.Equal(t, proof3, proof)

	proof, err = pp.GetProof(shardID, []byte("hash4"))
	require.Nil(t, err)
	require.Equal(t, proof4, proof)
	proof, err = pp.GetProofByNonce(4, shardID)
	require.Nil(t, err)
	require.Equal(t, proof4, proof)
}

func TestProofsPool_HasProofForDifferentHash(t *testing.T) {
	const nonce = uint64(7)
	pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)
	currentHash := []byte("current")
	competingHash := []byte("competing")
	require.True(t, pp.AddProof(&block.HeaderProof{
		HeaderHash:    currentHash,
		HeaderNonce:   nonce,
		HeaderShardId: shardID,
	}))

	require.False(t, pp.HasProofForDifferentHash(shardID, nonce, currentHash))
	require.False(t, pp.HasProofForDifferentHash(shardID, nonce+1, currentHash))
	require.False(t, pp.HasProofForDifferentHash(shardID+1, nonce, currentHash))

	require.True(t, pp.AddProof(&block.HeaderProof{
		HeaderHash:    competingHash,
		HeaderNonce:   nonce,
		HeaderShardId: shardID,
	}))
	require.True(t, pp.HasProofForDifferentHash(shardID, nonce, currentHash))
	require.True(t, pp.HasProofForDifferentHash(shardID, nonce, competingHash))

	result := false
	allocations := testing.AllocsPerRun(100, func() {
		result = pp.HasProofForDifferentHash(shardID, nonce, currentHash)
	})
	require.True(t, result)
	require.Zero(t, allocations)
}

func TestProofsPool_Upsert(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

	ok := pp.UpsertProof(nil)
	require.False(t, ok)

	ok = pp.UpsertProof(proof1)
	require.True(t, ok)

	proof, err := pp.GetProof(shardID, []byte("hash1"))
	require.Nil(t, err)
	require.NotNil(t, proof)

	require.Equal(t, proof1.GetAggregatedSignature(), proof.GetAggregatedSignature())
	require.Equal(t, proof1.GetPubKeysBitmap(), proof.GetPubKeysBitmap())

	newProof1 := &block.HeaderProof{
		PubKeysBitmap:       []byte("newpubKeysBitmap1"),
		AggregatedSignature: []byte("newaggSig1"),
		HeaderHash:          []byte("hash1"),
		HeaderEpoch:         1,
		HeaderNonce:         1,
		HeaderShardId:       shardID,
	}

	ok = pp.UpsertProof(newProof1)
	require.True(t, ok)

	proof, err = pp.GetProof(shardID, []byte("hash1"))
	require.Nil(t, err)
	require.NotNil(t, proof)

	require.Equal(t, newProof1.GetAggregatedSignature(), proof.GetAggregatedSignature())
	require.Equal(t, newProof1.GetPubKeysBitmap(), proof.GetPubKeysBitmap())
}

func TestProofsPool_UpsertMultipleHashes(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(3, 10)

	// Upsert 10 different-hash proofs for the same nonce, increasing rounds.
	// They are kept as competing proofs, capped at the max per nonce (lowest rounds win).
	for i := 0; i < 10; i++ {
		proof := &block.HeaderProof{
			HeaderHash:    []byte{byte(i)}, // Different hash each time
			HeaderNonce:   5,               // Same nonce
			HeaderRound:   uint64(10 + i),
			HeaderShardId: shardID,
		}
		ok := pp.UpsertProof(proof)
		require.True(t, ok, "upsert %d should succeed", i)
	}

	// The lowest-round proofs are retained up to the cap
	for i := 0; i < proofscache.MaxProofsPerNonce; i++ {
		require.True(t, pp.HasProof(shardID, []byte{byte(i)}), "low round hash[%d] should be retained", i)
	}

	// The higher-round proofs beyond the cap are evicted
	for i := proofscache.MaxProofsPerNonce; i < 10; i++ {
		require.False(t, pp.HasProof(shardID, []byte{byte(i)}), "high round hash[%d] should be evicted", i)
	}

	// The canonical proof at nonce 5 is the lowest-round one
	proofByNonce, err := pp.GetProofByNonce(5, shardID)
	require.Nil(t, err)
	require.Equal(t, []byte{0}, proofByNonce.GetHeaderHash(), "nonce 5 should map to the lowest round hash")

	proofs, err := pp.GetProofsByNonce(5, shardID)
	require.Nil(t, err)
	require.Equal(t, proofscache.MaxProofsPerNonce, len(proofs))
}

func TestProofsPool_IsProofEqual(t *testing.T) {
	t.Parallel()

	t.Run("not existing proof, should fail", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		ok := pp.IsProofInPoolEqualTo(proof1)
		require.False(t, ok)
	})

	t.Run("nil provided proof, should fail", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		ok := pp.IsProofInPoolEqualTo(nil)
		require.False(t, ok)
	})

	t.Run("same proof, should return true", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		ok := pp.UpsertProof(proof1)
		require.True(t, ok)

		ok = pp.IsProofInPoolEqualTo(proof1)
		require.True(t, ok)
	})

	t.Run("not equal, should return false", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		ok := pp.UpsertProof(proof1)
		require.True(t, ok)

		newProof1 := &block.HeaderProof{
			PubKeysBitmap:       []byte("newpubKeysBitmap1"),
			AggregatedSignature: []byte("newaggSig1"),
			HeaderHash:          []byte("hash1"),
			HeaderEpoch:         1,
			HeaderNonce:         1,
			HeaderShardId:       shardID,
		}

		ok = pp.IsProofInPoolEqualTo(newProof1)
		require.False(t, ok)
	})
}

func TestProofsPool_RegisterHandler(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

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

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		_ = pp.AddProof(proof1)
		_ = pp.AddProof(proof2)
		_ = pp.AddProof(proof3)
		_ = pp.AddProof(proof4)

		_, err := pp.GetProof(shardID, []byte("hash2"))
		require.Nil(t, err)
		_, err = pp.GetProof(shardID, []byte("hash3"))
		require.Nil(t, err)
		_, err = pp.GetProof(shardID, []byte("hash4"))
		require.Nil(t, err)
	})

	t.Run("should not cleanup if nonce smaller or equal to delta", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

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

	pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

	numOperations := 1000

	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	cnt := uint32(0)

	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			switch idx % 9 {
			case 0, 1, 2:
				_ = pp.AddProof(generateProof())
			case 3:
				_, err := pp.GetProof(generateRandomShardID(), generateRandomHash())
				if errors.Is(err, proofscache.ErrMissingProof) {
					atomic.AddUint32(&cnt, 1)
				}
			case 4:
				_, _ = pp.GetProofByNonce(generateRandomNonce(100), generateRandomShardID())
			case 5:
				_ = pp.CleanupProofsBehindNonce(generateRandomShardID(), generateRandomNonce(100))
			case 6:
				handler := func(proof data.HeaderProofHandler) {
				}
				pp.RegisterHandler(handler)
			case 7:
				_, _ = pp.GetProofsByNonce(generateRandomNonce(100), generateRandomShardID())
			case 8:
				handler := func(proof data.HeaderProofHandler, competingProofs []data.HeaderProofHandler) {
				}
				pp.RegisterEquivocationHandler(handler)
			default:
				assert.Fail(t, "should have not beed called")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

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

func TestProofsPool_CompetingProofsAtSameNonce(t *testing.T) {
	t.Parallel()

	proofRound6 := &block.HeaderProof{
		HeaderHash:    []byte("hashA"),
		HeaderNonce:   5,
		HeaderRound:   6,
		HeaderShardId: shardID,
	}
	proofRound7 := &block.HeaderProof{
		HeaderHash:    []byte("hashB"),
		HeaderNonce:   5,
		HeaderRound:   7,
		HeaderShardId: shardID,
	}

	t.Run("both proofs retained, canonical is lowest round regardless of add order", func(t *testing.T) {
		t.Parallel()

		for _, proofs := range [][]*block.HeaderProof{
			{proofRound6, proofRound7},
			{proofRound7, proofRound6},
		} {
			pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)
			require.True(t, pp.AddProof(proofs[0]))
			require.True(t, pp.AddProof(proofs[1]))

			require.True(t, pp.HasProof(shardID, []byte("hashA")))
			require.True(t, pp.HasProof(shardID, []byte("hashB")))

			proof, err := pp.GetProofByNonce(5, shardID)
			require.Nil(t, err)
			require.Equal(t, proofRound6, proof)

			allProofs, err := pp.GetProofsByNonce(5, shardID)
			require.Nil(t, err)
			require.Equal(t, 2, len(allProofs))
			require.Equal(t, proofRound6, allProofs[0])
			require.Equal(t, proofRound7, allProofs[1])
		}
	})

	t.Run("same round ties break on lowest hash", func(t *testing.T) {
		t.Parallel()

		tieHigh := &block.HeaderProof{HeaderHash: []byte("hashZ"), HeaderNonce: 5, HeaderRound: 6, HeaderShardId: shardID}

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)
		require.True(t, pp.AddProof(tieHigh))
		require.True(t, pp.AddProof(proofRound6))

		proof, err := pp.GetProofByNonce(5, shardID)
		require.Nil(t, err)
		require.Equal(t, proofRound6, proof)
	})

	t.Run("equivocation handler notified with competing proofs", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		type equivocationEvent struct {
			newProof  data.HeaderProofHandler
			competing []data.HeaderProofHandler
		}
		eventChan := make(chan equivocationEvent, 2)
		pp.RegisterEquivocationHandler(nil)
		pp.RegisterEquivocationHandler(func(headerProof data.HeaderProofHandler, competingProofs []data.HeaderProofHandler) {
			eventChan <- equivocationEvent{newProof: headerProof, competing: competingProofs}
		})

		require.True(t, pp.AddProof(proofRound6))
		select {
		case <-eventChan:
			require.Fail(t, "must not notify on the first proof at a nonce")
		case <-time.After(50 * time.Millisecond):
		}

		require.True(t, pp.AddProof(proofRound7))
		select {
		case event := <-eventChan:
			require.Equal(t, proofRound7, event.newProof)
			require.Equal(t, []data.HeaderProofHandler{proofRound6}, event.competing)
		case <-time.After(time.Second):
			require.Fail(t, "equivocation handler was not notified")
		}

		// re-adding an already stored proof at the equivocated nonce must not re-fire the event
		require.True(t, pp.UpsertProof(proofRound7))
		select {
		case <-eventChan:
			require.Fail(t, "must not notify again on a same-hash re-add")
		case <-time.After(50 * time.Millisecond):
		}
	})

	t.Run("no notification for different nonces or same hash re-add", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		notified := make(chan struct{}, 2)
		pp.RegisterEquivocationHandler(func(_ data.HeaderProofHandler, _ []data.HeaderProofHandler) {
			notified <- struct{}{}
		})

		require.True(t, pp.AddProof(proof1))
		require.True(t, pp.AddProof(proof2))
		require.False(t, pp.AddProof(proof1))
		require.True(t, pp.UpsertProof(proof1))

		select {
		case <-notified:
			require.Fail(t, "must not notify without a different-hash proof at the same nonce")
		case <-time.After(50 * time.Millisecond):
		}
	})

	t.Run("cleanup removes all proofs at the nonce", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)
		require.True(t, pp.AddProof(proofRound6))
		require.True(t, pp.AddProof(proofRound7))

		err := pp.CleanupProofsBehindNonce(shardID, 5+cleanupDelta+1)
		require.Nil(t, err)

		require.False(t, pp.HasProof(shardID, []byte("hashA")))
		require.False(t, pp.HasProof(shardID, []byte("hashB")))
		_, err = pp.GetProofsByNonce(5, shardID)
		require.NotNil(t, err)
	})
}

func TestProofsPool_GetProofsByNonce_Missing(t *testing.T) {
	t.Parallel()

	pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

	_, err := pp.GetProofsByNonce(5, shardID)
	require.NotNil(t, err)

	_ = pp.AddProof(proof1)
	_, err = pp.GetProofsByNonce(5, shardID)
	require.NotNil(t, err)
}

func TestProofsPool_AddProofIfNoneAtNonce(t *testing.T) {
	t.Parallel()

	t.Run("nil proof should not be added", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		added, existing := pp.AddProofIfNoneAtNonce(nil)
		require.False(t, added)
		require.Nil(t, existing)
	})

	t.Run("should add on free nonce and notify subscribers", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		notifyChan := make(chan data.HeaderProofHandler, 2)
		pp.RegisterHandler(func(headerProof data.HeaderProofHandler) {
			notifyChan <- headerProof
		})

		added, existing := pp.AddProofIfNoneAtNonce(proof1)
		require.True(t, added)
		require.Nil(t, existing)

		proof, err := pp.GetProofByNonce(proof1.GetHeaderNonce(), shardID)
		require.Nil(t, err)
		require.Equal(t, proof1, proof)

		select {
		case notified := <-notifyChan:
			require.Equal(t, proof1, notified)
		case <-time.After(time.Second):
			require.Fail(t, "subscriber was not notified on add")
		}

		// rejected adds must not notify
		added, _ = pp.AddProofIfNoneAtNonce(proof1)
		require.False(t, added)
		select {
		case <-notifyChan:
			require.Fail(t, "subscriber must not be notified on a rejected add")
		case <-time.After(50 * time.Millisecond):
		}
	})

	t.Run("should reject same hash at nonce", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		_, _ = pp.AddProofIfNoneAtNonce(proof1)
		added, existing := pp.AddProofIfNoneAtNonce(proof1)
		require.False(t, added)
		require.Equal(t, proof1, existing)
	})

	t.Run("should reject different hash at nonce without eviction", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		competingProof := &block.HeaderProof{
			PubKeysBitmap:       []byte("pubKeysBitmap2"),
			AggregatedSignature: []byte("aggSig2"),
			HeaderHash:          []byte("competing hash"),
			HeaderEpoch:         1,
			HeaderNonce:         proof1.GetHeaderNonce(),
			HeaderShardId:       shardID,
		}

		_, _ = pp.AddProofIfNoneAtNonce(proof1)
		added, existing := pp.AddProofIfNoneAtNonce(competingProof)
		require.False(t, added)
		require.Equal(t, proof1, existing)

		// the first proof is still reachable both by nonce and by hash
		proof, err := pp.GetProofByNonce(proof1.GetHeaderNonce(), shardID)
		require.Nil(t, err)
		require.Equal(t, proof1, proof)

		proof, err = pp.GetProof(shardID, proof1.GetHeaderHash())
		require.Nil(t, err)
		require.Equal(t, proof1, proof)

		require.False(t, pp.HasProof(shardID, competingProof.GetHeaderHash()))
	})

	t.Run("concurrent adds at the same nonce should admit exactly one", func(t *testing.T) {
		t.Parallel()

		pp := proofscache.NewProofsPool(cleanupDelta, bucketSize)

		numConcurrent := 100
		var numAdded uint32
		var wg sync.WaitGroup
		wg.Add(numConcurrent)
		for i := 0; i < numConcurrent; i++ {
			go func(idx int) {
				defer wg.Done()

				proof := &block.HeaderProof{
					HeaderHash:    []byte(fmt.Sprintf("hash_%d", idx)),
					HeaderNonce:   42,
					HeaderShardId: shardID,
				}
				added, _ := pp.AddProofIfNoneAtNonce(proof)
				if added {
					atomic.AddUint32(&numAdded, 1)
				}
			}(i)
		}
		wg.Wait()

		require.Equal(t, uint32(1), numAdded)

		_, err := pp.GetProofByNonce(42, shardID)
		require.Nil(t, err)
	})
}
