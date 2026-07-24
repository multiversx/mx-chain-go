package track

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

// IsSettledCrossHeader returns true if a proofed child extending the header is known locally, from
// the tracked headers or from the headers pool; only META headers settle this way, since a proofed
// shard child does not exclude a lower-round sibling that gathers one too
func (bbt *baseBlockTrack) IsSettledCrossHeader(header data.HeaderHandler, headerHash []byte) bool {
	if check.IfNil(header) || len(headerHash) == 0 {
		return false
	}

	shardID := header.GetShardID()
	if shardID != core.MetachainShardId {
		return false
	}

	childNonce := header.GetNonce() + 1

	trackedChildren, trackedChildrenHashes := bbt.GetTrackedHeadersWithNonce(shardID, childNonce)
	if bbt.holdProofedChild(trackedChildren, trackedChildrenHashes, headerHash, shardID) {
		return true
	}

	pooledChildren, pooledChildrenHashes, err := bbt.headersPool.GetHeadersByNonceAndShardId(childNonce, shardID)
	if err != nil {
		return false
	}

	return bbt.holdProofedChild(pooledChildren, pooledChildrenHashes, headerHash, shardID)
}

func (bbt *baseBlockTrack) holdProofedChild(
	children []data.HeaderHandler,
	childrenHashes [][]byte,
	parentHash []byte,
	shardID uint32,
) bool {
	for i, child := range children {
		if check.IfNil(child) || !bytes.Equal(child.GetPrevHash(), parentHash) {
			continue
		}

		if bbt.proofsPool.HasProof(shardID, childrenHashes[i]) {
			return true
		}
	}

	return false
}
