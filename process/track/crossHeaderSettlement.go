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
		return isMetaHeaderHeldFinalWithEvidence(
			bbt.proofsPool,
			header,
			headerHash,
			bbt.getTrackedOrPooledMetaHeader,
			bbt.hasTrackedOrPooledMetaDescendants,
		)
	}

	childNonce := header.GetNonce() + 1

	trackedChildren, trackedChildrenHashes := bbt.GetTrackedHeadersWithNonce(shardID, childNonce)
	if holdProofedChild(bbt.proofsPool, trackedChildren, trackedChildrenHashes, headerHash, shardID) {
		return true
	}

	return hasProofedChildInPool(bbt.headersPool, bbt.proofsPool, shardID, headerHash, childNonce)
}

func (bbt *baseBlockTrack) getTrackedOrPooledMetaHeader(headerHash []byte, nonce uint64) data.HeaderHandler {
	header := getMetaHeaderFromPool(bbt.headersPool, headerHash, nonce)
	if !check.IfNil(header) {
		return header
	}

	trackedHeaders, trackedHashes := bbt.GetTrackedHeadersWithNonce(core.MetachainShardId, nonce)
	for index, trackedHeader := range trackedHeaders {
		if index >= len(trackedHashes) || !bytes.Equal(trackedHashes[index], headerHash) || check.IfNil(trackedHeader) ||
			trackedHeader.GetShardID() != core.MetachainShardId || trackedHeader.GetNonce() != nonce {
			continue
		}

		return trackedHeader
	}

	return nil
}

func (bbt *baseBlockTrack) hasTrackedOrPooledMetaDescendants(
	nonce uint64,
	parentHash []byte,
	depth int,
) bool {
	pooledHeaders, pooledHashes, err := bbt.headersPool.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err == nil && bbt.hasProofedMetaDescendant(pooledHeaders, pooledHashes, nonce, parentHash, depth) {
		return true
	}

	trackedHeaders, trackedHashes := bbt.GetTrackedHeadersWithNonce(core.MetachainShardId, nonce)
	return bbt.hasProofedMetaDescendant(trackedHeaders, trackedHashes, nonce, parentHash, depth)
}

func (bbt *baseBlockTrack) hasProofedMetaDescendant(
	headers []data.HeaderHandler,
	hashes [][]byte,
	nonce uint64,
	parentHash []byte,
	depth int,
) bool {
	for index, header := range headers {
		if index >= len(hashes) || check.IfNil(header) || header.GetShardID() != core.MetachainShardId ||
			header.GetNonce() != nonce || !bytes.Equal(header.GetPrevHash(), parentHash) ||
			!bbt.proofsPool.HasProof(core.MetachainShardId, hashes[index]) {
			continue
		}
		if depth <= 1 || bbt.hasTrackedOrPooledMetaDescendants(nonce+1, hashes[index], depth-1) {
			return true
		}
	}

	return false
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
