package sync

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	testscommonDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

func TestShardSettlementChecker_IsSettled(t *testing.T) {
	t.Parallel()

	nonce := uint64(10)
	headerHash := []byte("headerHash")

	t.Run("defers to the meta finality view with the self shard id and the cross-notarized anchor", func(t *testing.T) {
		t.Parallel()

		var gotShardID uint32
		var gotHash []byte
		var gotNonce, gotAnchor uint64
		checker := &shardSettlementChecker{
			selfShardID: 3,
			blockTracker: &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
					require.Equal(t, core.MetachainShardId, shardID)
					return &block.MetaBlock{Nonce: 42}, []byte("metaHash"), nil
				},
			},
			metaFinalityView: &testscommon.MetaFinalityViewStub{
				IsIncludedInHeldFinalMetaBlockCalled: func(shardID uint32, hash []byte, hdrNonce uint64, anchor uint64) bool {
					gotShardID, gotHash, gotNonce, gotAnchor = shardID, hash, hdrNonce, anchor
					return true
				},
			},
		}

		require.True(t, checker.isSettled(nonce, headerHash))
		require.Equal(t, uint32(3), gotShardID)
		require.Equal(t, headerHash, gotHash)
		require.Equal(t, nonce, gotNonce)
		require.Equal(t, uint64(42), gotAnchor)
	})

	t.Run("a failing tracker degrades to anchor zero", func(t *testing.T) {
		t.Parallel()

		var gotAnchor uint64
		checker := &shardSettlementChecker{
			selfShardID: 0,
			blockTracker: &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return nil, nil, errors.New("tracker error")
				},
			},
			metaFinalityView: &testscommon.MetaFinalityViewStub{
				IsIncludedInHeldFinalMetaBlockCalled: func(_ uint32, _ []byte, _ uint64, anchor uint64) bool {
					gotAnchor = anchor
					return false
				},
			},
		}

		require.False(t, checker.isSettled(nonce, headerHash))
		require.Equal(t, uint64(0), gotAnchor)
	})

	t.Run("a proofed shard child alone does not settle", func(t *testing.T) {
		t.Parallel()

		checker := &shardSettlementChecker{
			selfShardID:      0,
			blockTracker:     &mock.BlockTrackerMock{},
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

		// signing locks per round, not per nonce, so both siblings can gather a proofed child; only depth-2 counts
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

func TestShardSettlementChecker_DeadCrossNotarizedMeta(t *testing.T) {
	t.Parallel()

	metaNonce := uint64(30)
	metaHash := []byte("crossNotarizedMetaHash")
	crossNotarizedMeta := &block.MetaBlock{Nonce: metaNonce}

	t.Run("failing tracker reports nothing", func(t *testing.T) {
		t.Parallel()

		checker := &shardSettlementChecker{
			blockTracker: &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return nil, nil, errors.New("tracker error")
				},
			},
			metaFinalityView: &testscommon.MetaFinalityViewStub{
				IsDeadMetaBlockCalled: func(_ []byte, _ uint64) bool {
					require.Fail(t, "the view must not be consulted without a tracker verdict")
					return false
				},
			},
		}

		_, _, isDead := checker.deadCrossNotarizedMeta()
		require.False(t, isDead)
	})

	t.Run("defers to the shared dead-branch evidence with the pointer hash and nonce", func(t *testing.T) {
		t.Parallel()

		var gotHash []byte
		var gotNonce uint64
		checker := &shardSettlementChecker{
			blockTracker: &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return crossNotarizedMeta, metaHash, nil
				},
			},
			metaFinalityView: &testscommon.MetaFinalityViewStub{
				IsDeadMetaBlockCalled: func(headerHash []byte, nonce uint64) bool {
					gotHash, gotNonce = headerHash, nonce
					return true
				},
			},
		}

		deadMeta, deadHash, isDead := checker.deadCrossNotarizedMeta()
		require.True(t, isDead)
		require.Equal(t, crossNotarizedMeta, deadMeta)
		require.Equal(t, metaHash, deadHash)
		require.Equal(t, metaHash, gotHash)
		require.Equal(t, metaNonce, gotNonce)
	})

	t.Run("a live pointer reports nothing", func(t *testing.T) {
		t.Parallel()

		checker := &shardSettlementChecker{
			blockTracker: &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return crossNotarizedMeta, metaHash, nil
				},
			},
			metaFinalityView: &testscommon.MetaFinalityViewStub{},
		}

		_, _, isDead := checker.deadCrossNotarizedMeta()
		require.False(t, isDead)
	})
}
