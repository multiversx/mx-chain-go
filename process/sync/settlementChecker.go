package sync

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
)

// inclusionScanSpan is the per-round budget of the resumable authority scan, matching the view's
// own window bound
const inclusionScanSpan = 16

// shardSettlementChecker settles shard blocks from the canonical settlement-ready meta view.
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

func (checker *shardSettlementChecker) settlementVerdict(
	nonce uint64,
	localHash []byte,
	competitorHash []byte,
	scanFrom uint64,
	scanTo uint64,
) (bool, bool) {
	anchor, continuation, continuationHashes, ok := checker.canonicalMetaView()
	if !ok {
		return false, false
	}

	competitorSettled := checker.isSettledInMetaView(
		anchor, continuation, continuationHashes, nonce, competitorHash, scanFrom, scanTo)
	localSettled := checker.isSettledInMetaView(
		anchor, continuation, continuationHashes, nonce, localHash, scanFrom, scanTo)

	return localSettled, competitorSettled
}

func (checker *shardSettlementChecker) isSettledInMetaView(
	anchor metaHeaderWithHash,
	continuation []data.HeaderHandler,
	continuationHashes [][]byte,
	nonce uint64,
	headerHash []byte,
	scanFrom uint64,
	scanTo uint64,
) bool {
	if len(headerHash) == 0 {
		return false
	}

	if checker.isSettlementAuthorityForShardHeader(anchor.header, anchor.hash, nonce, headerHash) {
		return true
	}
	for index, metaHeader := range continuation {
		if index >= len(continuationHashes) || metaHeader.GetNonce() < scanFrom || metaHeader.GetNonce() > scanTo {
			continue
		}
		meta, isMeta := metaHeader.(data.MetaHeaderHandler)
		if isMeta && checker.isSettlementAuthorityForShardHeader(meta, continuationHashes[index], nonce, headerHash) {
			return true
		}
	}

	return false
}

type metaHeaderWithHash struct {
	header data.MetaHeaderHandler
	hash   []byte
}

func (checker *shardSettlementChecker) canonicalMetaView() (
	metaHeaderWithHash,
	[]data.HeaderHandler,
	[][]byte,
	bool,
) {
	metaAnchor, metaAnchorHash, err := checker.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil || check.IfNil(metaAnchor) || len(metaAnchorHash) == 0 ||
		checker.metaFinalityView.IsDeadMetaBlock(metaAnchorHash, metaAnchor.GetNonce()) {
		return metaHeaderWithHash{}, nil, nil, false
	}
	anchorHandler, ok := metaAnchor.(data.MetaHeaderHandler)
	if !ok {
		return metaHeaderWithHash{}, nil, nil, false
	}

	continuation, hashes := checker.blockTracker.ComputeLongestChain(core.MetachainShardId, metaAnchor)
	for index, header := range continuation {
		if index >= len(hashes) || check.IfNil(header) ||
			checker.metaFinalityView.IsDeadMetaBlock(hashes[index], header.GetNonce()) {
			continuation = continuation[:index]
			hashes = hashes[:index]
			break
		}
	}

	return metaHeaderWithHash{header: anchorHandler, hash: metaAnchorHash}, continuation, hashes, true
}

func (checker *shardSettlementChecker) isSettlementAuthorityForShardHeader(
	metaHeader data.MetaHeaderHandler,
	metaHash []byte,
	shardNonce uint64,
	shardHash []byte,
) bool {
	return checker.metaFinalityView.IsMetaHeaderSettlementReady(metaHeader, metaHash) &&
		checker.metaFinalityView.IsShardHeaderIncluded(metaHeader, checker.selfShardID, shardHash, shardNonce)
}

func (checker *shardSettlementChecker) resolveNotarizedHeader(
	nonce uint64,
	candidates []notarizedHeaderCandidate,
) []byte {
	if len(candidates) < 2 {
		return nil
	}

	anchor, continuation, continuationHashes, ok := checker.canonicalMetaView()
	if !ok {
		return nil
	}

	var selectedHash []byte
	var unique bool
	if checker.metaFinalityView.IsMetaHeaderSettlementReady(anchor.header, anchor.hash) {
		selectedHash, unique = checker.selectIncludedCandidate(anchor.header, nonce, candidates, nil)
		if !unique {
			return nil
		}
	}

	for index, header := range continuation {
		metaHeader, ok := header.(data.MetaHeaderHandler)
		if !ok || check.IfNil(metaHeader) || index >= len(continuationHashes) ||
			!checker.metaFinalityView.IsMetaHeaderSettlementReady(metaHeader, continuationHashes[index]) {
			continue
		}

		selectedHash, unique = checker.selectIncludedCandidate(metaHeader, nonce, candidates, selectedHash)
		if !unique {
			return nil
		}
	}

	return selectedHash
}

func (checker *shardSettlementChecker) selectIncludedCandidate(
	metaHeader data.MetaHeaderHandler,
	nonce uint64,
	candidates []notarizedHeaderCandidate,
	selectedHash []byte,
) ([]byte, bool) {
	foundDirectReference := false
	for _, shardInfo := range process.GetShardHeadersReferencedByMeta(metaHeader) {
		if shardInfo.GetShardID() != checker.selfShardID || shardInfo.GetNonce() != nonce ||
			!containsNotarizedCandidate(candidates, nonce, shardInfo.GetHeaderHash()) {
			continue
		}

		foundDirectReference = true
		if len(selectedHash) > 0 && !bytes.Equal(selectedHash, shardInfo.GetHeaderHash()) {
			return nil, false
		}
		selectedHash = append(selectedHash[:0], shardInfo.GetHeaderHash()...)
	}
	if foundDirectReference {
		return selectedHash, true
	}

	for _, candidate := range candidates {
		if candidate.nonce != nonce || !checker.metaFinalityView.IsShardHeaderIncluded(
			metaHeader,
			checker.selfShardID,
			candidate.hash,
			candidate.nonce,
		) {
			continue
		}

		if len(selectedHash) > 0 && !bytes.Equal(selectedHash, candidate.hash) {
			return nil, false
		}
		selectedHash = append(selectedHash[:0], candidate.hash...)
	}

	return selectedHash, true
}

func containsNotarizedCandidate(candidates []notarizedHeaderCandidate, nonce uint64, hash []byte) bool {
	for _, candidate := range candidates {
		if candidate.nonce == nonce && bytes.Equal(candidate.hash, hash) {
			return true
		}
	}

	return false
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
	return track.HasMetaReconciliationEvidence(checker.headers, checker.proofs, nonce, headerHash)
}

func (checker *metaSettlementChecker) settlementVerdict(
	nonce uint64,
	localHash []byte,
	competitorHash []byte,
	_ uint64,
	_ uint64,
) (bool, bool) {
	competitorSettled := len(competitorHash) > 0 && checker.isSettled(nonce, competitorHash, 0, 0)
	localSettled := checker.isSettled(nonce, localHash, 0, 0)

	return localSettled, competitorSettled
}

func (checker *metaSettlementChecker) resolveNotarizedHeader(_ uint64, _ []notarizedHeaderCandidate) []byte {
	return nil
}

// deadCrossNotarizedMeta never reports on meta nodes; meta reconciles through the equivocation path
func (checker *metaSettlementChecker) deadCrossNotarizedMeta() (data.HeaderHandler, []byte, bool) {
	return nil, nil, false
}
