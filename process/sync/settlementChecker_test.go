package sync

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	testscommonDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

func TestShardSettlementChecker_IsSettled(t *testing.T) {
	t.Parallel()

	nonce := uint64(10)
	headerHash := []byte("headerHash")

	t.Run("defers to the meta finality view with the self shard id", func(t *testing.T) {
		t.Parallel()

		var gotShardID uint32
		var gotHash []byte
		var gotNonce uint64
		checker := &shardSettlementChecker{
			selfShardID: 3,
			metaFinalityView: &testscommon.MetaFinalityViewStub{
				IsIncludedInHeldFinalMetaBlockCalled: func(shardID uint32, hash []byte, hdrNonce uint64) bool {
					gotShardID, gotHash, gotNonce = shardID, hash, hdrNonce
					return true
				},
			},
		}

		require.True(t, checker.isSettled(nonce, headerHash))
		require.Equal(t, uint32(3), gotShardID)
		require.Equal(t, headerHash, gotHash)
		require.Equal(t, nonce, gotNonce)
	})

	t.Run("a proofed shard child alone does not settle", func(t *testing.T) {
		t.Parallel()

		checker := &shardSettlementChecker{
			selfShardID:      0,
			metaFinalityView: &testscommon.MetaFinalityViewStub{},
		}

		require.False(t, checker.isSettled(nonce, headerHash))
	})
}

type pooledHeader struct {
	header data.HeaderHandler
	hash   []byte
}

func newMetaCheckerWithPools(byNonce map[uint64][]pooledHeader, proofedHashes ...[]byte) *metaSettlementChecker {
	return &metaSettlementChecker{
		headers: &pool.HeadersPoolStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
				entries, ok := byNonce[hdrNonce]
				if !ok || shardID != core.MetachainShardId {
					return nil, nil, errors.New("no headers at nonce")
				}

				headers := make([]data.HeaderHandler, 0, len(entries))
				hashes := make([][]byte, 0, len(entries))
				for _, entry := range entries {
					headers = append(headers, entry.header)
					hashes = append(hashes, entry.hash)
				}

				return headers, hashes, nil
			},
		},
		proofs: &testscommonDataRetriever.ProofsPoolMock{
			HasProofCalled: func(_ uint32, hash []byte) bool {
				for _, proofed := range proofedHashes {
					if string(proofed) == string(hash) {
						return true
					}
				}
				return false
			},
		},
	}
}

func TestMetaSettlementChecker_IsSettled(t *testing.T) {
	t.Parallel()

	nonce := uint64(10)
	parentHash := []byte("parentHash")
	childHash := []byte("childHash")
	grandChildHash := []byte("grandChildHash")

	child := &block.MetaBlock{Nonce: nonce + 1, PrevHash: parentHash}
	grandChild := &block.MetaBlock{Nonce: nonce + 2, PrevHash: childHash}

	t.Run("a proofed child alone does not settle", func(t *testing.T) {
		t.Parallel()

		// under per-round R0 both siblings can gather a proofed child; only depth-2 counts
		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
		}, childHash)

		require.False(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("a proofed child with a proofed linked grandchild settles", func(t *testing.T) {
		t.Parallel()

		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, childHash, grandChildHash)

		require.True(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("an unproofed grandchild does not settle", func(t *testing.T) {
		t.Parallel()

		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, childHash)

		require.False(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("an unproofed child does not settle even with a proofed grandchild", func(t *testing.T) {
		t.Parallel()

		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, grandChildHash)

		require.False(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("a grandchild extending a sibling child does not settle", func(t *testing.T) {
		t.Parallel()

		strayGrandChild := &block.MetaBlock{Nonce: nonce + 2, PrevHash: []byte("siblingChildHash")}
		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{strayGrandChild, grandChildHash}},
		}, childHash, grandChildHash)

		require.False(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("a proofed child of a sibling does not settle", func(t *testing.T) {
		t.Parallel()

		strayChild := &block.MetaBlock{Nonce: nonce + 1, PrevHash: []byte("siblingHash")}
		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{strayChild, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, childHash, grandChildHash)

		require.False(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("only a proofed child with a proofed extension settles among siblings", func(t *testing.T) {
		t.Parallel()

		unproofedChildHash := []byte("unproofedChildHash")
		unproofedChild := &block.MetaBlock{Nonce: nonce + 1, PrevHash: parentHash}
		strandedGrandChild := &block.MetaBlock{Nonce: nonce + 2, PrevHash: unproofedChildHash}
		strandedGrandChildHash := []byte("strandedGrandChildHash")

		// the proofed grandchild extends the unproofed sibling, the proofed child has no extension
		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{unproofedChild, unproofedChildHash}, {child, childHash}},
			nonce + 2: {{strandedGrandChild, strandedGrandChildHash}},
		}, childHash, strandedGrandChildHash)
		require.False(t, checker.isSettled(nonce, parentHash))

		checker = newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{unproofedChild, unproofedChildHash}, {child, childHash}},
			nonce + 2: {{strandedGrandChild, strandedGrandChildHash}, {grandChild, grandChildHash}},
		}, childHash, strandedGrandChildHash, grandChildHash)
		require.True(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("no child known", func(t *testing.T) {
		t.Parallel()

		checker := &metaSettlementChecker{
			headers: &pool.HeadersPoolStub{},
			proofs:  &testscommonDataRetriever.ProofsPoolMock{},
		}

		require.False(t, checker.isSettled(nonce, parentHash))
	})
}
