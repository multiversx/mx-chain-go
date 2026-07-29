package sync

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// inclusionScanSpan is the per-round budget of the resumable authority scan, matching the view's
// own window bound
const inclusionScanSpan = 16

// shardSettlementChecker settles a shard block on the verdict of the settlement authority: a meta
// block the node holds final notarized it, or one of its descendants
type shardSettlementChecker struct {
	metaFinalityView process.MetaFinalityView
	blockTracker     process.BlockTracker
	headers          dataRetriever.HeadersPool
	proofs           dataRetriever.ProofsPool
	requestHandler   process.RequestHandler
	selfShardID      uint32
}

// prepareInclusionScan resumes the authority scan from the cursor, requesting missing headers and
// proofs paired, so successive rounds examine every nonce between the fork era anchor and the head;
// the cursor passes a nonce only once a proofed header there is extended by a proofed child, since
// a lone proofed sibling may be an equivocation loser hiding the still missing authority block
func (checker *shardSettlementChecker) prepareInclusionScan(scanCursor uint64) (uint64, uint64, uint64) {
	if scanCursor == 0 {
		// the cross-notarized meta is frozen at the fork era, anchoring the scan for any stranding
		lastCrossNotarizedMeta, _, err := checker.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
		if err != nil || check.IfNil(lastCrossNotarizedMeta) || lastCrossNotarizedMeta.GetNonce() == 0 {
			return 0, 0, 0
		}
		scanCursor = lastCrossNotarizedMeta.GetNonce()
	}

	scanFrom := scanCursor
	scanTo := scanFrom + inclusionScanSpan - 1
	poolHead := highestPooledMetaNonce(checker.headers)
	if scanTo > poolHead {
		scanTo = poolHead
	}
	if scanTo < scanFrom {
		return scanFrom, scanFrom - 1, scanCursor
	}

	nextCursor := scanCursor
	for nonce := scanFrom; nonce <= scanTo; nonce++ {
		if checker.hasWitnessedMetaHeaderAtNonce(nonce) {
			if nextCursor == nonce {
				nextCursor = nonce + 1
			}
			continue
		}

		isProofedPoolHead := nonce == poolHead && checker.hasProofedMetaHeaderAtNonce(nonce)
		if isProofedPoolHead {
			// no child could be pooled yet; the descending window of the inclusion check covers it
			continue
		}

		checker.requestHandler.RequestMetaHeaderByNonce(nonce)
		checker.requestHandler.RequestEquivalentProofByNonce(core.MetachainShardId, nonce)
	}

	return scanFrom, scanTo, nextCursor
}

// hasWitnessedMetaHeaderAtNonce requires a proofed header extended by a proofed child, the same
// evidence class the inclusion check itself consumes
func (checker *shardSettlementChecker) hasWitnessedMetaHeaderAtNonce(nonce uint64) bool {
	headers, hashes, err := checker.headers.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	for i, header := range headers {
		if check.IfNil(header) || !checker.proofs.HasProof(core.MetachainShardId, hashes[i]) {
			continue
		}
		if checker.hasProofedMetaChildOf(hashes[i], nonce+1) {
			return true
		}
	}

	return false
}

func (checker *shardSettlementChecker) hasProofedMetaChildOf(parentHash []byte, childNonce uint64) bool {
	children, childHashes, err := checker.headers.GetHeadersByNonceAndShardId(childNonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	for i, child := range children {
		if check.IfNil(child) || !bytes.Equal(child.GetPrevHash(), parentHash) {
			continue
		}
		if checker.proofs.HasProof(core.MetachainShardId, childHashes[i]) {
			return true
		}
	}

	return false
}

func (checker *shardSettlementChecker) hasProofedMetaHeaderAtNonce(nonce uint64) bool {
	_, hashes, err := checker.headers.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err != nil {
		return false
	}

	for _, hash := range hashes {
		if checker.proofs.HasProof(core.MetachainShardId, hash) {
			return true
		}
	}

	return false
}

func highestPooledMetaNonce(headers dataRetriever.HeadersPool) uint64 {
	highest := uint64(0)
	for _, nonce := range headers.Nonces(core.MetachainShardId) {
		if nonce > highest {
			highest = nonce
		}
	}

	return highest
}

func (checker *shardSettlementChecker) isSettled(nonce uint64, headerHash []byte, scanFrom uint64, scanTo uint64) bool {
	return checker.metaFinalityView.IsIncludedInHeldFinalMetaBlock(checker.selfShardID, headerHash, nonce, scanFrom, scanTo)
}

// deadCrossNotarizedMeta returns the last cross-notarized meta block when the authority provably
// built past it, per the shared dead-branch evidence of the meta finality view
func (checker *shardSettlementChecker) deadCrossNotarizedMeta() (data.HeaderHandler, []byte, bool) {
	lastCrossNotarizedMeta, lastCrossNotarizedHash, err := checker.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil || check.IfNil(lastCrossNotarizedMeta) {
		return nil, nil, false
	}

	if !checker.metaFinalityView.IsDeadMetaBlock(lastCrossNotarizedHash, lastCrossNotarizedMeta.GetNonce()) {
		return nil, nil, false
	}

	return lastCrossNotarizedMeta, lastCrossNotarizedHash, true
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

// prepareInclusionScan is a no-op: the meta settle test is a position-independent child lookup
func (checker *metaSettlementChecker) prepareInclusionScan(_ uint64) (uint64, uint64, uint64) {
	return 0, 0, 0
}

func (checker *metaSettlementChecker) isSettled(nonce uint64, headerHash []byte, _ uint64, _ uint64) bool {
	return checker.hasProofedDescendants(nonce+1, headerHash, metaSettledDescendantsDepth)
}

// deadCrossNotarizedMeta never reports on meta nodes; meta reconciles through the equivocation path
func (checker *metaSettlementChecker) deadCrossNotarizedMeta() (data.HeaderHandler, []byte, bool) {
	return nil, nil, false
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
