package track_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

const (
	metaParentNonce = uint64(5)
	metaParentRound = uint64(10)
)

var metaParentHash = []byte("metaParentHash")

func newFinalityViewWithPools(t *testing.T) (dataRetriever.PoolsHolder, process.MetaFinalityView) {
	pools := dataRetrieverMock.NewPoolsHolderMock()
	view, err := track.NewMetaFinalityView(track.ArgsMetaFinalityView{
		HeadersPool: pools.Headers(),
		ProofsPool:  pools.Proofs(),
	})
	require.Nil(t, err)

	return pools, view
}

func addMetaProof(t *testing.T, pools dataRetriever.PoolsHolder, hash []byte, nonce uint64, round uint64) {
	added := pools.Proofs().AddProof(&block.HeaderProof{
		HeaderHash:    hash,
		HeaderNonce:   nonce,
		HeaderRound:   round,
		HeaderShardId: core.MetachainShardId,
	})
	require.True(t, added)
}

func addProofedMetaParent(t *testing.T, pools dataRetriever.PoolsHolder) *block.MetaBlock {
	parent := &block.MetaBlock{Nonce: metaParentNonce, Round: metaParentRound}
	pools.Headers().AddHeader(metaParentHash, parent)
	addMetaProof(t, pools, metaParentHash, metaParentNonce, metaParentRound)

	return parent
}

func TestNewMetaFinalityView(t *testing.T) {
	t.Parallel()

	pools := dataRetrieverMock.NewPoolsHolderMock()

	t.Run("nil headers pool", func(t *testing.T) {
		t.Parallel()

		view, err := track.NewMetaFinalityView(track.ArgsMetaFinalityView{ProofsPool: pools.Proofs()})
		require.Nil(t, view)
		require.Equal(t, track.ErrNilHeadersPool, err)
	})

	t.Run("nil proofs pool", func(t *testing.T) {
		t.Parallel()

		view, err := track.NewMetaFinalityView(track.ArgsMetaFinalityView{HeadersPool: pools.Headers()})
		require.Nil(t, view)
		require.Equal(t, track.ErrNilProofsPool, err)
	})

	t.Run("all dependencies provided", func(t *testing.T) {
		t.Parallel()

		view, err := track.NewMetaFinalityView(track.ArgsMetaFinalityView{
			HeadersPool: pools.Headers(),
			ProofsPool:  pools.Proofs(),
		})
		require.Nil(t, err)
		require.False(t, view.IsInterfaceNil())
	})
}

func TestMetaFinalityView_IsMetaHeaderHeldFinal(t *testing.T) {
	t.Parallel()

	headerHash := []byte("headerHash")

	t.Run("nil header or empty hash", func(t *testing.T) {
		t.Parallel()

		_, view := newFinalityViewWithPools(t)

		require.False(t, view.IsMetaHeaderHeldFinal(nil, headerHash))
		require.False(t, view.IsMetaHeaderHeldFinal(&block.MetaBlock{Nonce: 6}, nil))
	})

	t.Run("shard header is never held final by this view", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addMetaProof(t, pools, headerHash, 6, metaParentRound+1)

		shardHeader := &block.Header{ShardID: 0, Nonce: 6, Round: metaParentRound + 1, PrevHash: metaParentHash}
		require.False(t, view.IsMetaHeaderHeldFinal(shardHeader, headerHash))
	})

	t.Run("unproofed header", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 1, PrevHash: metaParentHash}
		require.False(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("non contended over a proofed parent is instantly final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 1, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+1)

		require.True(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("non contended over an unproofed parent is not final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		pools.Headers().AddHeader(metaParentHash, &block.MetaBlock{Nonce: metaParentNonce, Round: metaParentRound})

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 1, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+1)

		require.False(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("missing parent is not final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 1, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+1)

		require.False(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("prev hash pointing at a non consecutive nonce is not final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		pools.Headers().AddHeader(metaParentHash, &block.MetaBlock{Nonce: metaParentNonce - 1, Round: metaParentRound})
		addMetaProof(t, pools, metaParentHash, metaParentNonce-1, metaParentRound)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 1, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+1)

		require.False(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("contended without a child is not final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 4, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+4)

		require.False(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("contended with an unproofed child is not final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 4, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+4)
		pools.Headers().AddHeader([]byte("childHash"), &block.MetaBlock{Nonce: 7, Round: metaParentRound + 5, PrevHash: headerHash})

		require.False(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("contended with a proofed child is settled", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 4, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+4)

		childHash := []byte("childHash")
		pools.Headers().AddHeader(childHash, &block.MetaBlock{Nonce: 7, Round: metaParentRound + 5, PrevHash: headerHash})
		addMetaProof(t, pools, childHash, 7, metaParentRound+5)

		require.True(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})

	t.Run("a proofed child of a sibling does not settle the header", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		header := &block.MetaBlock{Nonce: 6, Round: metaParentRound + 4, PrevHash: metaParentHash}
		addMetaProof(t, pools, headerHash, 6, metaParentRound+4)

		siblingChildHash := []byte("siblingChildHash")
		pools.Headers().AddHeader(siblingChildHash, &block.MetaBlock{Nonce: 7, Round: metaParentRound + 5, PrevHash: []byte("siblingHash")})
		addMetaProof(t, pools, siblingChildHash, 7, metaParentRound+5)

		require.False(t, view.IsMetaHeaderHeldFinal(header, headerHash))
	})
}

func TestMetaFinalityView_IsIncludedInHeldFinalMetaBlock(t *testing.T) {
	t.Parallel()

	const shardID = uint32(0)
	const shardNonce = uint64(20)

	shardHash := []byte("shardHash")
	metaHash := []byte("metaHash")

	addFinalMetaBlockAt := func(t *testing.T, pools dataRetriever.PoolsHolder, nonce uint64, referenced ...[]byte) {
		shardInfo := make([]block.ShardData, 0, len(referenced))
		for _, hash := range referenced {
			shardInfo = append(shardInfo, block.ShardData{HeaderHash: hash, ShardID: shardID})
		}

		metaBlock := &block.MetaBlock{
			Nonce:     nonce,
			Round:     metaParentRound + 1,
			PrevHash:  metaParentHash,
			ShardInfo: shardInfo,
		}
		pools.Headers().AddHeader(metaHash, metaBlock)
		addMetaProof(t, pools, metaHash, nonce, metaParentRound+1)
	}

	t.Run("empty hash", func(t *testing.T) {
		t.Parallel()

		_, view := newFinalityViewWithPools(t)

		require.False(t, view.IsIncludedInHeldFinalMetaBlock(shardID, nil, shardNonce))
	})

	t.Run("empty meta pool", func(t *testing.T) {
		t.Parallel()

		_, view := newFinalityViewWithPools(t)

		require.False(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})

	t.Run("exact hash referenced by a held final meta block", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)
		addFinalMetaBlockAt(t, pools, metaParentNonce+1, shardHash)

		require.True(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})

	t.Run("descendant at depth two referenced by a held final meta block", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		childHash := []byte("shardChildHash")
		grandChildHash := []byte("shardGrandChildHash")
		pools.Headers().AddHeader(childHash, &block.Header{ShardID: shardID, Nonce: shardNonce + 1, PrevHash: shardHash})
		pools.Headers().AddHeader(grandChildHash, &block.Header{ShardID: shardID, Nonce: shardNonce + 2, PrevHash: childHash})

		addFinalMetaBlockAt(t, pools, metaParentNonce+1, grandChildHash)

		require.True(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})

	t.Run("descendant past the walk bound is not reached", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		prevHash := shardHash
		var lastHash []byte
		for depth := uint64(1); depth <= track.MaxOwnDescendantsScannedForInclusion+1; depth++ {
			lastHash = []byte(fmt.Sprintf("shardDescendant%d", depth))
			pools.Headers().AddHeader(lastHash, &block.Header{ShardID: shardID, Nonce: shardNonce + depth, PrevHash: prevHash})
			prevHash = lastHash
		}

		addFinalMetaBlockAt(t, pools, metaParentNonce+1, lastHash)

		require.False(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})

	t.Run("a header off the branch is not a descendant", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		strangerHash := []byte("strangerHash")
		pools.Headers().AddHeader(strangerHash, &block.Header{ShardID: shardID, Nonce: shardNonce + 1, PrevHash: []byte("otherParent")})

		addFinalMetaBlockAt(t, pools, metaParentNonce+1, strangerHash)

		require.False(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})

	t.Run("referenced for another shard", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)
		addFinalMetaBlockAt(t, pools, metaParentNonce+1, shardHash)

		require.False(t, view.IsIncludedInHeldFinalMetaBlock(shardID+1, shardHash, shardNonce))
	})

	t.Run("referencing meta block is not held final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)

		// contended and childless, so the notarization carries no authority yet
		metaBlock := &block.MetaBlock{
			Nonce:     metaParentNonce + 1,
			Round:     metaParentRound + 4,
			PrevHash:  metaParentHash,
			ShardInfo: []block.ShardData{{HeaderHash: shardHash, ShardID: shardID}},
		}
		pools.Headers().AddHeader(metaHash, metaBlock)
		addMetaProof(t, pools, metaHash, metaParentNonce+1, metaParentRound+4)

		require.False(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})

	t.Run("referencing meta block below the scan window", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)
		addFinalMetaBlockAt(t, pools, metaParentNonce+1, shardHash)

		aheadNonce := metaParentNonce + 1 + 2*track.MaxMetaBlocksScannedForInclusion
		pools.Headers().AddHeader([]byte("aheadHash"), &block.MetaBlock{Nonce: aheadNonce, Round: metaParentRound + 100})

		require.False(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})

	t.Run("v3 meta block references through the shard info proposal", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		pools.Headers().AddHeader(metaParentHash, &block.MetaBlockV3{Nonce: metaParentNonce, Round: metaParentRound})
		addMetaProof(t, pools, metaParentHash, metaParentNonce, metaParentRound)

		metaBlock := &block.MetaBlockV3{
			Nonce:             metaParentNonce + 1,
			Round:             metaParentRound + 1,
			PrevHash:          metaParentHash,
			ShardInfoProposal: []block.ShardDataProposal{{HeaderHash: shardHash, ShardID: shardID}},
		}
		pools.Headers().AddHeader(metaHash, metaBlock)
		addMetaProof(t, pools, metaHash, metaParentNonce+1, metaParentRound+1)

		require.True(t, view.IsIncludedInHeldFinalMetaBlock(shardID, shardHash, shardNonce))
	})
}

func TestMetaFinalityView_HasHeldFinalCompetitorAtNonce(t *testing.T) {
	t.Parallel()

	const contendedNonce = uint64(6)

	ownHash := []byte("ownHash")
	competitorHash := []byte("competitorHash")

	// the node's own header: contended and childless, so it is not held final by itself
	newOwnHeader := func() *block.MetaBlock {
		return &block.MetaBlock{Nonce: contendedNonce, Round: metaParentRound + 4, PrevHash: metaParentHash}
	}

	t.Run("nil header or empty hash", func(t *testing.T) {
		t.Parallel()

		_, view := newFinalityViewWithPools(t)

		require.False(t, view.HasHeldFinalCompetitorAtNonce(nil, ownHash))
		require.False(t, view.HasHeldFinalCompetitorAtNonce(newOwnHeader(), nil))
	})

	t.Run("a single proof at the nonce short circuits without touching the headers pool", func(t *testing.T) {
		t.Parallel()

		pools := dataRetrieverMock.NewPoolsHolderMock()
		failingHeadersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
				require.Fail(t, "should not have read the headers pool")
				return nil, nil
			},
			GetHeaderByNonceAndShardIdCalled: func(_ uint64, _ uint32) ([]data.HeaderHandler, [][]byte, error) {
				require.Fail(t, "should not have read the headers pool")
				return nil, nil, nil
			},
			NoncesCalled: func(_ uint32) []uint64 {
				require.Fail(t, "should not have read the headers pool")
				return nil
			},
		}

		view, err := track.NewMetaFinalityView(track.ArgsMetaFinalityView{
			HeadersPool: failingHeadersPool,
			ProofsPool:  pools.Proofs(),
		})
		require.Nil(t, err)

		addMetaProof(t, pools, ownHash, contendedNonce, metaParentRound+4)

		require.False(t, view.HasHeldFinalCompetitorAtNonce(newOwnHeader(), ownHash))
	})

	t.Run("held final competitor while the own header is not final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)
		addMetaProof(t, pools, ownHash, contendedNonce, metaParentRound+4)

		// the lower round sibling is non contended, so it is instantly final over the proofed parent
		pools.Headers().AddHeader(competitorHash, &block.MetaBlock{Nonce: contendedNonce, Round: metaParentRound + 1, PrevHash: metaParentHash})
		addMetaProof(t, pools, competitorHash, contendedNonce, metaParentRound+1)

		require.True(t, view.HasHeldFinalCompetitorAtNonce(newOwnHeader(), ownHash))
	})

	t.Run("both siblings held final yields no verdict", func(t *testing.T) {
		t.Parallel()

		// the accepted fact-2 residual; a non exclusive verdict here would make two nodes revert
		// onto each other's branch forever
		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)
		addMetaProof(t, pools, ownHash, contendedNonce, metaParentRound+4)

		ownChildHash := []byte("ownChildHash")
		pools.Headers().AddHeader(ownChildHash, &block.MetaBlock{Nonce: contendedNonce + 1, Round: metaParentRound + 5, PrevHash: ownHash})
		addMetaProof(t, pools, ownChildHash, contendedNonce+1, metaParentRound+5)

		pools.Headers().AddHeader(competitorHash, &block.MetaBlock{Nonce: contendedNonce, Round: metaParentRound + 1, PrevHash: metaParentHash})
		addMetaProof(t, pools, competitorHash, contendedNonce, metaParentRound+1)

		require.False(t, view.HasHeldFinalCompetitorAtNonce(newOwnHeader(), ownHash))
	})

	t.Run("competitor header missing from the pool", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)
		addMetaProof(t, pools, ownHash, contendedNonce, metaParentRound+4)
		addMetaProof(t, pools, competitorHash, contendedNonce, metaParentRound+1)

		require.False(t, view.HasHeldFinalCompetitorAtNonce(newOwnHeader(), ownHash))
	})

	t.Run("competitor is not held final", func(t *testing.T) {
		t.Parallel()

		pools, view := newFinalityViewWithPools(t)
		addProofedMetaParent(t, pools)
		addMetaProof(t, pools, ownHash, contendedNonce, metaParentRound+4)

		// contended and childless as well, so neither side carries a verdict
		pools.Headers().AddHeader(competitorHash, &block.MetaBlock{Nonce: contendedNonce, Round: metaParentRound + 3, PrevHash: metaParentHash})
		addMetaProof(t, pools, competitorHash, contendedNonce, metaParentRound+3)

		require.False(t, view.HasHeldFinalCompetitorAtNonce(newOwnHeader(), ownHash))
	})
}
