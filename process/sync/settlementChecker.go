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

// metaSettlementChecker settles a meta block on the depth-1 settle-on-child rule; meta has no
// external authority to defer to
type metaSettlementChecker struct {
	headers dataRetriever.HeadersPool
	proofs  dataRetriever.ProofsPool
}

func (checker *metaSettlementChecker) isSettled(nonce uint64, headerHash []byte) bool {
	return checker.hasProofedChild(nonce+1, headerHash)
}

func (checker *metaSettlementChecker) hasProofedChild(nonce uint64, parentHash []byte) bool {
	headers, hashes, err := checker.headers.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	for i, header := range headers {
		if check.IfNil(header) || !bytes.Equal(header.GetPrevHash(), parentHash) {
			continue
		}
		if checker.proofs.HasProof(core.MetachainShardId, hashes[i]) {
			return true
		}
	}

	return false
}
