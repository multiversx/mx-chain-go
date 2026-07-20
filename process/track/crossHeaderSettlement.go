package track

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

// IsSettledCrossHeader returns true if the header can no longer change: a proofed child extending
// it is known locally, from the tracked headers or from the headers pool
func (bbt *baseBlockTrack) IsSettledCrossHeader(header data.HeaderHandler, headerHash []byte) bool {
	if check.IfNil(header) || len(headerHash) == 0 {
		return false
	}

	shardID := header.GetShardID()
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
