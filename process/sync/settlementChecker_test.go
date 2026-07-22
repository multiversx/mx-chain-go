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

func TestMetaSettlementChecker_IsSettled(t *testing.T) {
	t.Parallel()

	nonce := uint64(10)
	parentHash := []byte("parentHash")
	childHash := []byte("childHash")

	newChecker := func(child data.HeaderHandler, proofedHashes ...[]byte) *metaSettlementChecker {
		return &metaSettlementChecker{
			headers: &pool.HeadersPoolStub{
				GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
					if hdrNonce != nonce+1 || shardID != core.MetachainShardId {
						return nil, nil, errors.New("no headers at nonce")
					}
					return []data.HeaderHandler{child}, [][]byte{childHash}, nil
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

	t.Run("a proofed child settles the meta header", func(t *testing.T) {
		t.Parallel()

		child := &block.MetaBlock{Nonce: nonce + 1, PrevHash: parentHash}
		checker := newChecker(child, childHash)

		require.True(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("an unproofed child does not settle", func(t *testing.T) {
		t.Parallel()

		child := &block.MetaBlock{Nonce: nonce + 1, PrevHash: parentHash}
		checker := newChecker(child)

		require.False(t, checker.isSettled(nonce, parentHash))
	})

	t.Run("a proofed child of a sibling does not settle", func(t *testing.T) {
		t.Parallel()

		child := &block.MetaBlock{Nonce: nonce + 1, PrevHash: []byte("siblingHash")}
		checker := newChecker(child, childHash)

		require.False(t, checker.isSettled(nonce, parentHash))
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
