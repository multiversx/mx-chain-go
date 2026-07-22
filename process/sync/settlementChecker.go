package sync

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// shardSettlementChecker settles a shard block on the verdict of the settlement authority: a meta
// block the node holds final notarized it, or one of its descendants
type shardSettlementChecker struct {
	metaFinalityView process.MetaFinalityView
	selfShardID      uint32
}

func (checker *shardSettlementChecker) isSettled(nonce uint64, headerHash []byte) bool {
	return checker.metaFinalityView.IsIncludedInHeldFinalMetaBlock(checker.selfShardID, headerHash, nonce)
}

// metaSettledDescendantsDepth requires a proofed child that is itself extended by a proofed child;
// depth 2 closes the depth-1 double-extension corner, the depth-2 residual is accepted
const metaSettledDescendantsDepth = 2

// metaSettlementChecker settles a meta block on a fully proofed descendant chain; meta has no
// external authority to defer to
type metaSettlementChecker struct {
	headers dataRetriever.HeadersPool
	proofs  dataRetriever.ProofsPool
}

func (checker *metaSettlementChecker) isSettled(nonce uint64, headerHash []byte) bool {
	return checker.hasProofedDescendants(nonce+1, headerHash, metaSettledDescendantsDepth)
}

// hasProofedDescendants reports whether a chain of the given depth, proofed at every level,
// extends parentHash
func (checker *metaSettlementChecker) hasProofedDescendants(nonce uint64, parentHash []byte, depth int) bool {
	headers, hashes, err := checker.headers.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	for i, header := range headers {
		if check.IfNil(header) || !bytes.Equal(header.GetPrevHash(), parentHash) {
			continue
		}
		if !checker.proofs.HasProof(core.MetachainShardId, hashes[i]) {
			continue
		}
		if depth <= 1 || checker.hasProofedDescendants(nonce+1, hashes[i], depth-1) {
			return true
		}
	}

	return false
}
