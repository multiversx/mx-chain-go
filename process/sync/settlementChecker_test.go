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

	t.Run("defers to the meta finality view with the self shard id and the scan window", func(t *testing.T) {
		t.Parallel()

		var gotShardID uint32
		var gotHash []byte
		var gotNonce, gotFrom, gotTo uint64
		checker := &shardSettlementChecker{
			selfShardID: 3,
			metaFinalityView: &testscommon.MetaFinalityViewStub{
				IsIncludedInHeldFinalMetaBlockCalled: func(shardID uint32, hash []byte, hdrNonce uint64, from uint64, to uint64) bool {
					gotShardID, gotHash, gotNonce, gotFrom, gotTo = shardID, hash, hdrNonce, from, to
					return true
				},
			},
		}

		require.True(t, checker.isSettled(nonce, headerHash, 42, 57))
		require.Equal(t, uint32(3), gotShardID)
		require.Equal(t, headerHash, gotHash)
		require.Equal(t, nonce, gotNonce)
		require.Equal(t, uint64(42), gotFrom)
		require.Equal(t, uint64(57), gotTo)
	})

	t.Run("a proofed shard child alone does not settle", func(t *testing.T) {
		t.Parallel()

		checker := &shardSettlementChecker{
			selfShardID:      0,
			blockTracker:     &mock.BlockTrackerMock{},
			metaFinalityView: &testscommon.MetaFinalityViewStub{},
		}

		require.False(t, checker.isSettled(nonce, headerHash, 0, 0))
	})
}

type scanRequests struct {
	headerNonces [][]uint64
	proofNonces  [][]uint64
}

func newScanChecker(
	anchorNonce uint64,
	trackerErr error,
	pooledNonces map[uint64][]pooledHeader,
	proofedHashes map[string]bool,
	requests *scanRequests,
) *shardSettlementChecker {
	return &shardSettlementChecker{
		selfShardID: 0,
		blockTracker: &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
				if trackerErr != nil {
					return nil, nil, trackerErr
				}
				return &block.MetaBlock{Nonce: anchorNonce}, []byte("metaHash"), nil
			},
		},
		metaFinalityView: &testscommon.MetaFinalityViewStub{},
		headers: &pool.HeadersPoolStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
				entries, ok := pooledNonces[hdrNonce]
				if !ok {
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
			NoncesCalled: func(shardID uint32) []uint64 {
				nonces := make([]uint64, 0, len(pooledNonces))
				for pooledNonce := range pooledNonces {
					nonces = append(nonces, pooledNonce)
				}
				return nonces
			},
		},
		proofs: &testscommonDataRetriever.ProofsPoolMock{
			HasProofCalled: func(_ uint32, hash []byte) bool {
				return proofedHashes[string(hash)]
			},
		},
		requestHandler: &testscommon.RequestHandlerStub{
			RequestMetaHeaderByNonceCalled: func(hdrNonce uint64) {
				requests.headerNonces = append(requests.headerNonces, []uint64{hdrNonce})
			},
			RequestEquivalentProofByNonceCalled: func(_ uint32, hdrNonce uint64) {
				requests.proofNonces = append(requests.proofNonces, []uint64{hdrNonce})
			},
		},
	}
}

func TestShardSettlementChecker_PrepareInclusionScan(t *testing.T) {
	t.Parallel()

	anchor := uint64(40)
	// a linked chain, one header per nonce, every hash proofed
	proofedRun := func(from, to uint64) (map[uint64][]pooledHeader, map[string]bool) {
		pooled := make(map[uint64][]pooledHeader)
		proofed := make(map[string]bool)
		for nonce := from; nonce <= to; nonce++ {
			hash := []byte{byte(nonce)}
			header := &block.MetaBlock{Nonce: nonce, PrevHash: []byte{byte(nonce - 1)}}
			pooled[nonce] = []pooledHeader{{header: header, hash: hash}}
			proofed[string(hash)] = true
		}
		return pooled, proofed
	}

	t.Run("zero cursor initializes at the fork era anchor", func(t *testing.T) {
		t.Parallel()

		requests := &scanRequests{}
		pooled, proofed := proofedRun(anchor, anchor+100)
		checker := newScanChecker(anchor, nil, pooled, proofed, requests)

		scanFrom, scanTo, nextCursor := checker.prepareInclusionScan(0)
		require.Equal(t, anchor, scanFrom)
		require.Equal(t, anchor+inclusionScanSpan-1, scanTo)
		require.Equal(t, anchor+inclusionScanSpan, nextCursor)
		require.Empty(t, requests.headerNonces)
	})

	t.Run("a failing tracker holds the scan instead of scanning from zero", func(t *testing.T) {
		t.Parallel()

		requests := &scanRequests{}
		pooled, proofed := proofedRun(anchor, anchor+100)
		checker := newScanChecker(anchor, errors.New("tracker error"), pooled, proofed, requests)

		scanFrom, scanTo, nextCursor := checker.prepareInclusionScan(0)
		require.Equal(t, uint64(0), scanFrom)
		require.Equal(t, uint64(0), scanTo)
		require.Equal(t, uint64(0), nextCursor)
		require.Empty(t, requests.headerNonces)
	})

	t.Run("the cursor advances only past nonces witnessed by a proofed child and missing data is requested paired", func(t *testing.T) {
		t.Parallel()

		requests := &scanRequests{}
		pooled, proofed := proofedRun(anchor, anchor+2)
		// nonce anchor+3 pooled and linked but UNPROOFED: it does not witness anchor+2
		pooled[anchor+3] = []pooledHeader{{header: &block.MetaBlock{Nonce: anchor + 3, PrevHash: []byte{byte(anchor + 2)}}, hash: []byte{byte(anchor + 3)}}}
		pooled[anchor+100] = []pooledHeader{{header: &block.MetaBlock{Nonce: anchor + 100}, hash: []byte{200}}} // pool head far above
		checker := newScanChecker(anchor, nil, pooled, proofed, requests)

		scanFrom, scanTo, nextCursor := checker.prepareInclusionScan(anchor)
		require.Equal(t, anchor, scanFrom)
		require.Equal(t, anchor+inclusionScanSpan-1, scanTo)
		require.Equal(t, anchor+2, nextCursor)

		// every unwitnessed nonce in the window is requested, header and proof paired
		require.Equal(t, len(requests.headerNonces), len(requests.proofNonces))
		require.Equal(t, int(inclusionScanSpan)-2, len(requests.headerNonces))
		require.Contains(t, requests.headerNonces, []uint64{anchor + 2})
	})

	t.Run("a lone proofed sibling does not advance the cursor and its nonce keeps being requested", func(t *testing.T) {
		t.Parallel()

		requests := &scanRequests{}
		// a proofed decoy at the anchor: no pooled child links back to it, so the canonical
		// sibling may still be missing and the nonce is not complete evidence
		decoyHash := []byte("decoyHash")
		pooled := map[uint64][]pooledHeader{
			anchor:     {{header: &block.MetaBlock{Nonce: anchor, PrevHash: []byte("forkParent")}, hash: decoyHash}},
			anchor + 1: {{header: &block.MetaBlock{Nonce: anchor + 1, PrevHash: []byte("missingSibling")}, hash: []byte("childHash")}},
		}
		proofed := map[string]bool{string(decoyHash): true, "childHash": true}
		checker := newScanChecker(anchor, nil, pooled, proofed, requests)

		_, _, nextCursor := checker.prepareInclusionScan(anchor)
		require.Equal(t, anchor, nextCursor)
		require.Contains(t, requests.headerNonces, []uint64{anchor})
		require.Contains(t, requests.proofNonces, []uint64{anchor})
	})

	t.Run("a proofed header below a pool gap is still requested", func(t *testing.T) {
		t.Parallel()

		requests := &scanRequests{}
		// only the true pool head may go unrequested; a gap right above a proofed header can
		// hide a missing sibling there just as well
		pooled := map[uint64][]pooledHeader{
			anchor:     {{header: &block.MetaBlock{Nonce: anchor, PrevHash: []byte("forkParent")}, hash: []byte("decoyHash")}},
			anchor + 5: {{header: &block.MetaBlock{Nonce: anchor + 5}, hash: []byte("aboveGapHash")}},
		}
		proofed := map[string]bool{"decoyHash": true, "aboveGapHash": true}
		checker := newScanChecker(anchor, nil, pooled, proofed, requests)

		_, _, nextCursor := checker.prepareInclusionScan(anchor)
		require.Equal(t, anchor, nextCursor)
		require.Contains(t, requests.headerNonces, []uint64{anchor})
		require.NotContains(t, requests.headerNonces, []uint64{anchor + 5})
	})

	t.Run("the window is capped at the pool head and the childless proofed head is not requested", func(t *testing.T) {
		t.Parallel()

		requests := &scanRequests{}
		pooled, proofed := proofedRun(anchor, anchor+3)
		checker := newScanChecker(anchor, nil, pooled, proofed, requests)

		scanFrom, scanTo, nextCursor := checker.prepareInclusionScan(anchor)
		require.Equal(t, anchor, scanFrom)
		require.Equal(t, anchor+3, scanTo)
		// the head cannot have a pooled child yet; the cursor waits there and the descending
		// window of the inclusion check covers it
		require.Equal(t, anchor+3, nextCursor)
		require.Empty(t, requests.headerNonces)
	})

	t.Run("a cursor above the pool head yields an empty window without requests", func(t *testing.T) {
		t.Parallel()

		requests := &scanRequests{}
		pooled, proofed := proofedRun(anchor, anchor+3)
		checker := newScanChecker(anchor, nil, pooled, proofed, requests)

		scanFrom, scanTo, nextCursor := checker.prepareInclusionScan(anchor + 10)
		require.Equal(t, anchor+10, scanFrom)
		require.Greater(t, scanFrom, scanTo)
		require.Equal(t, anchor+10, nextCursor)
		require.Empty(t, requests.headerNonces)
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

		require.False(t, checker.isSettled(nonce, parentHash, 0, 0))
	})

	t.Run("a proofed child with a proofed linked grandchild settles", func(t *testing.T) {
		t.Parallel()

		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, childHash, grandChildHash)

		require.True(t, checker.isSettled(nonce, parentHash, 0, 0))
	})

	t.Run("an unproofed grandchild does not settle", func(t *testing.T) {
		t.Parallel()

		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, childHash)

		require.False(t, checker.isSettled(nonce, parentHash, 0, 0))
	})

	t.Run("an unproofed child does not settle even with a proofed grandchild", func(t *testing.T) {
		t.Parallel()

		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, grandChildHash)

		require.False(t, checker.isSettled(nonce, parentHash, 0, 0))
	})

	t.Run("a grandchild extending a sibling child does not settle", func(t *testing.T) {
		t.Parallel()

		strayGrandChild := &block.MetaBlock{Nonce: nonce + 2, PrevHash: []byte("siblingChildHash")}
		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{child, childHash}},
			nonce + 2: {{strayGrandChild, grandChildHash}},
		}, childHash, grandChildHash)

		require.False(t, checker.isSettled(nonce, parentHash, 0, 0))
	})

	t.Run("a proofed child of a sibling does not settle", func(t *testing.T) {
		t.Parallel()

		strayChild := &block.MetaBlock{Nonce: nonce + 1, PrevHash: []byte("siblingHash")}
		checker := newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{strayChild, childHash}},
			nonce + 2: {{grandChild, grandChildHash}},
		}, childHash, grandChildHash)

		require.False(t, checker.isSettled(nonce, parentHash, 0, 0))
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
		require.False(t, checker.isSettled(nonce, parentHash, 0, 0))

		checker = newMetaCheckerWithPools(map[uint64][]pooledHeader{
			nonce + 1: {{unproofedChild, unproofedChildHash}, {child, childHash}},
			nonce + 2: {{strandedGrandChild, strandedGrandChildHash}, {grandChild, grandChildHash}},
		}, childHash, strandedGrandChildHash, grandChildHash)
		require.True(t, checker.isSettled(nonce, parentHash, 0, 0))
	})

	t.Run("no child known", func(t *testing.T) {
		t.Parallel()

		checker := &metaSettlementChecker{
			headers: &pool.HeadersPoolStub{},
			proofs:  &testscommonDataRetriever.ProofsPoolMock{},
		}

		require.False(t, checker.isSettled(nonce, parentHash, 0, 0))
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
