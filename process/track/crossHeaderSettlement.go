package track

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// IsSettledCrossHeader uses the held-final rule for V3 meta headers and the proofed-child rule for
// legacy meta headers.
func (bbt *baseBlockTrack) IsSettledCrossHeader(header data.HeaderHandler, headerHash []byte) bool {
	if check.IfNil(header) || len(headerHash) == 0 {
		return false
	}

	shardID := header.GetShardID()
	if shardID != core.MetachainShardId {
		return false
	}
	if header.IsHeaderV3() {
		return isMetaHeaderHeldFinal(bbt.headersPool, bbt.proofsPool, header, headerHash)
	}

	childNonce := header.GetNonce() + 1

	trackedChildren, trackedChildrenHashes := bbt.GetTrackedHeadersWithNonce(shardID, childNonce)
	if holdProofedChild(bbt.proofsPool, trackedChildren, trackedChildrenHashes, headerHash, shardID) {
		return true
	}

	return hasProofedChildInPool(bbt.headersPool, bbt.proofsPool, shardID, headerHash, childNonce)
}

func hasProofedChildInPool(
	headersPool dataRetriever.HeadersPool,
	proofsPool dataRetriever.ProofsPool,
	shardID uint32,
	parentHash []byte,
	childNonce uint64,
) bool {
	children, childrenHashes, err := headersPool.GetHeadersByNonceAndShardId(childNonce, shardID)
	if err != nil {
		return false
	}

	return holdProofedChild(proofsPool, children, childrenHashes, parentHash, shardID)
}

func holdProofedChild(
	proofsPool dataRetriever.ProofsPool,
	children []data.HeaderHandler,
	childrenHashes [][]byte,
	parentHash []byte,
	shardID uint32,
) bool {
	for i, child := range children {
		if check.IfNil(child) || !bytes.Equal(child.GetPrevHash(), parentHash) {
			continue
		}

		if proofsPool.HasProof(shardID, childrenHashes[i]) {
			return true
		}
	}

	return false
}
