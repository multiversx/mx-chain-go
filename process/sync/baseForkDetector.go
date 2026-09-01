package sync

import (
	"bytes"
	"math"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/process"
)

type headerInfo struct {
	epoch    uint32
	nonce    uint64
	round    uint64
	hash     []byte
	prevHash []byte
	state    process.BlockHeaderState
	hasProof bool
}

type appendHeaderInfoResult struct {
	inserted bool
	enriched bool
}

type checkpointInfo struct {
	nonce uint64
	round uint64
	hash  []byte
}

type forkInfo struct {
	checkpoint              []*checkpointInfo
	finalCheckpoint         *checkpointInfo
	settledCheckpoint       *checkpointInfo
	probableHighestNonce    uint64
	highestNonceReceived    uint64
	rollBackNonce           uint64
	lastRoundWithForcedFork int64
}

// baseForkDetector defines a struct with necessary data needed for fork detection
type baseForkDetector struct {
	roundHandler consensus.RoundHandler

	headers                       map[uint64][]*headerInfo
	mutHeaders                    sync.RWMutex
	fork                          forkInfo
	mutFork                       sync.RWMutex
	mutProbableHighestNonceUpdate sync.Mutex

	shardID                uint32
	blackListHandler       process.TimeCacher
	genesisTime            int64
	supernovaGenesisTime   int64
	blockTracker           process.BlockTracker
	forkDetector           forkDetector
	genesisNonce           uint64
	genesisRound           uint64
	maxForkHeaderEpoch     uint32
	genesisEpoch           uint32
	enableEpochsHandler    common.EnableEpochsHandler
	enableRoundsHandler    common.EnableRoundsHandler
	proofsPool             process.ProofsPool
	chainParametersHandler common.ChainParametersHandler
	processConfigsHandler  common.ProcessConfigsHandler
}

// SetRollBackNonce sets the nonce where the chain should roll back
func (bfd *baseForkDetector) SetRollBackNonce(nonce uint64) {
	bfd.mutFork.Lock()
	bfd.fork.rollBackNonce = nonce
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) getRollBackNonce() uint64 {
	bfd.mutFork.RLock()
	nonce := bfd.fork.rollBackNonce
	bfd.mutFork.RUnlock()

	return nonce
}

func (bfd *baseForkDetector) setLastRoundWithForcedFork(round int64) {
	bfd.mutFork.Lock()
	bfd.fork.lastRoundWithForcedFork = round
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) lastRoundWithForcedFork() int64 {
	bfd.mutFork.RLock()
	round := bfd.fork.lastRoundWithForcedFork
	bfd.mutFork.RUnlock()

	return round
}

func (bfd *baseForkDetector) removePastOrInvalidRecords() {
	bfd.removePastHeaders()
	bfd.removeInvalidReceivedHeaders()
	bfd.removePastCheckpoints()
}

func (bfd *baseForkDetector) checkBlockBasicValidity(
	header data.HeaderHandler,
	headerHash []byte,
) error {

	if check.IfNil(header) {
		return ErrNilHeader
	}
	if headerHash == nil {
		return ErrNilHash
	}

	roundDif := int64(header.GetRound()) - int64(bfd.finalCheckpoint().round)
	nonceDif := int64(header.GetNonce()) - int64(bfd.finalCheckpoint().nonce)
	// TODO: Analyze if the acceptance of some headers which came for the next round could generate some attack vectors
	// bound against the round the current time falls into, not the stored index: a node slow to
	// advance its chronology must still record the headers it needs to catch up
	nextRound := bfd.roundHandler.IndexForCurrentTime() + 1

	bfd.blackListHandler.Sweep()
	if bfd.blackListHandler.Has(string(header.GetPrevHash())) {
		process.AddHeaderToBlackList(bfd.blackListHandler, headerHash)
		return process.ErrHeaderIsBlackListed
	}
	// TODO: This check could be removed when this protection mechanism would be implemented on interceptors side

	err := bfd.checkGenesisTimeForHeader(header)
	if err != nil {
		process.AddHeaderToBlackList(bfd.blackListHandler, headerHash)
		return ErrGenesisTimeMissmatch
	}
	if roundDif < 0 {
		return ErrLowerRoundInBlock
	}
	if nonceDif < 0 {
		return ErrLowerNonceInBlock
	}
	if int64(header.GetRound()) > nextRound {
		return ErrHigherRoundInBlock
	}
	if roundDif < nonceDif {
		return ErrHigherNonceInBlock
	}

	return nil
}

// removePastHeaders retains entries down to the settled checkpoint, so instantly finalized blocks
// keep their processed entries until the meta notarization arrives and settles them
func (bfd *baseForkDetector) removePastHeaders() {
	settledCheckpointNonce := bfd.settledCheckpoint().nonce

	bfd.mutHeaders.Lock()
	for nonce := range bfd.headers {
		if nonce < settledCheckpointNonce {
			delete(bfd.headers, nonce)
		}
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *baseForkDetector) removeInvalidReceivedHeaders() {
	finalCheckpointRound := bfd.finalCheckpoint().round
	finalCheckpointNonce := bfd.finalCheckpoint().nonce

	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		validHdrInfos := make([]*headerInfo, 0)
		for i := 0; i < len(hdrInfos); i++ {
			roundDif := int64(hdrInfos[i].round) - int64(finalCheckpointRound)
			nonceDif := int64(hdrInfos[i].nonce) - int64(finalCheckpointNonce)
			hasStateReceived := hdrInfos[i].state == process.BHReceived || hdrInfos[i].state == process.BHReceivedTooLate
			isReceivedHeaderInvalid := hasStateReceived && roundDif < nonceDif
			if isReceivedHeaderInvalid {
				continue
			}

			validHdrInfos = append(validHdrInfos, hdrInfos[i])
		}
		if len(validHdrInfos) == 0 {
			delete(bfd.headers, nonce)
			continue
		}

		bfd.headers[nonce] = validHdrInfos
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *baseForkDetector) removePastCheckpoints() {
	bfd.removeCheckpointsBehindNonce(bfd.finalCheckpoint().nonce)
}

func (bfd *baseForkDetector) removeCheckpointsBehindNonce(nonce uint64) {
	bfd.mutFork.Lock()
	preservedCheckpoint := make([]*checkpointInfo, 0)

	for i := 0; i < len(bfd.fork.checkpoint); i++ {
		if bfd.fork.checkpoint[i].nonce < nonce {
			continue
		}

		preservedCheckpoint = append(preservedCheckpoint, bfd.fork.checkpoint[i])
	}

	bfd.fork.checkpoint = preservedCheckpoint
	bfd.mutFork.Unlock()
}

// computeProbableHighestNonce computes the probable highest nonce from the valid received/processed headers
func (bfd *baseForkDetector) computeProbableHighestNonce() uint64 {
	finalCheckpoint := bfd.finalCheckpoint()
	settledCheckpoint := bfd.settledCheckpoint()
	lastCheckpoint := bfd.lastCheckpoint()
	probableHighestNonce := finalCheckpoint.nonce
	settledNonce := uint64(0)
	if settledCheckpoint != nil {
		settledNonce = settledCheckpoint.nonce
	}
	requiresBranchSelection := false

	bfd.mutHeaders.RLock()
	for nonce, headers := range bfd.headers {
		hasBranchSelectionEvidence, hasActionableHeader := bfd.classifyProbableHeaders(
			nonce,
			headers,
			finalCheckpoint,
			settledNonce,
		)
		if nonce >= finalCheckpoint.nonce && hasBranchSelectionEvidence {
			requiresBranchSelection = true
		}

		if nonce > probableHighestNonce && hasActionableHeader {
			probableHighestNonce = nonce
		}
	}

	if requiresBranchSelection && probableHighestNonce > finalCheckpoint.nonce {
		probableHighestNonce = bfd.computeBranchAwareProbable(finalCheckpoint, lastCheckpoint, probableHighestNonce)
	}
	bfd.mutHeaders.RUnlock()

	return probableHighestNonce
}

func (bfd *baseForkDetector) classifyProbableHeaders(
	nonce uint64,
	hdrInfos []*headerInfo,
	finalCheckpoint *checkpointInfo,
	settledNonce uint64,
) (bool, bool) {
	canSelectBranch := len(finalCheckpoint.hash) > 0
	hasBranchSelectionEvidence := false
	hasActionableHeader := false

	uniqueProofs := 0
	for index, hdrInfo := range hdrInfos {
		isV3 := false
		if canSelectBranch || hdrInfo.state == process.BHNotarized {
			isV3 = bfd.isAsyncExecutionEnabled(hdrInfo)
		}
		if hdrInfo.hasProof || hdrInfo.state == process.BHNotarized && isV3 {
			hasActionableHeader = true
		}
		if !canSelectBranch || !isV3 {
			continue
		}
		if hdrInfo.state == process.BHNotarized && nonce > settledNonce {
			hasBranchSelectionEvidence = true
		}
		if !hdrInfo.hasProof || bfd.hasEarlierSameHash(hdrInfos, index) {
			continue
		}
		if nonce == finalCheckpoint.nonce && !bytes.Equal(hdrInfo.hash, finalCheckpoint.hash) {
			hasBranchSelectionEvidence = true
		}
		if nonce > finalCheckpoint.nonce && nonce-finalCheckpoint.nonce == 1 &&
			len(hdrInfo.prevHash) > 0 && !bytes.Equal(hdrInfo.prevHash, finalCheckpoint.hash) {
			hasBranchSelectionEvidence = true
		}

		uniqueProofs++
		if uniqueProofs > 1 {
			hasBranchSelectionEvidence = true
		}
	}

	return hasBranchSelectionEvidence, hasActionableHeader
}

func (bfd *baseForkDetector) computeBranchAwareProbable(
	finalCheckpoint *checkpointInfo,
	lastCheckpoint *checkpointInfo,
	rawProbable uint64,
) uint64 {
	selectedHash := finalCheckpoint.hash
	actionableNonce := finalCheckpoint.nonce

	for nonce := finalCheckpoint.nonce + 1; ; nonce++ {
		hdrInfos := bfd.headers[nonce]
		notarizedHeader, numNotarizedHashes := bfd.getUniqueNotarizedHeader(hdrInfos)
		if numNotarizedHashes > 1 {
			return rawProbable
		}

		var selectedHeader *headerInfo
		if numNotarizedHashes == 1 {
			if !bfd.isAsyncExecutionEnabled(notarizedHeader) || len(notarizedHeader.prevHash) == 0 ||
				!bytes.Equal(notarizedHeader.prevHash, selectedHash) {
				return rawProbable
			}

			selectedHeader = notarizedHeader
		} else {
			if !bfd.hasProvenHeader(hdrInfos) || !bfd.allProvenHeadersHaveKnownV3Ancestry(hdrInfos) {
				return rawProbable
			}

			var differentEpochs bool
			selectedHeader, differentEpochs = bfd.getPreferredProvenChild(hdrInfos, selectedHash)
			if differentEpochs {
				return rawProbable
			}
			if selectedHeader == nil {
				if nonce <= lastCheckpoint.nonce {
					return rawProbable
				}
				if bfd.isCompleteLosingSuffix(nonce, rawProbable) {
					return actionableNonce
				}

				return rawProbable
			}
		}

		if nonce <= lastCheckpoint.nonce && !bfd.processedHeaderMatches(hdrInfos, selectedHeader.hash) {
			return rawProbable
		}

		selectedHash = selectedHeader.hash
		actionableNonce = nonce
		if nonce == rawProbable {
			return actionableNonce
		}
	}
}

func (bfd *baseForkDetector) getUniqueNotarizedHeader(hdrInfos []*headerInfo) (*headerInfo, int) {
	var notarizedHeader *headerInfo
	numNotarizedHashes := 0

	for index, hdrInfo := range hdrInfos {
		if hdrInfo.state != process.BHNotarized || bfd.hasEarlierSameHashWithState(hdrInfos, index) {
			continue
		}

		notarizedHeader = hdrInfo
		numNotarizedHashes++
	}

	return notarizedHeader, numNotarizedHashes
}

func (bfd *baseForkDetector) getPreferredProvenChild(hdrInfos []*headerInfo, parentHash []byte) (*headerInfo, bool) {
	var preferred *headerInfo

	for index, hdrInfo := range hdrInfos {
		if !hdrInfo.hasProof || bfd.hasEarlierSameHash(hdrInfos, index) || !bytes.Equal(hdrInfo.prevHash, parentHash) {
			continue
		}
		if preferred != nil && preferred.epoch != hdrInfo.epoch {
			return nil, true
		}
		if preferred == nil || isLowerRoundOrHash(hdrInfo.round, hdrInfo.hash, preferred.round, preferred.hash) {
			preferred = hdrInfo
		}
	}

	return preferred, false
}

func (bfd *baseForkDetector) processedHeaderMatches(hdrInfos []*headerInfo, selectedHash []byte) bool {
	var processedHash []byte

	for index, hdrInfo := range hdrInfos {
		if hdrInfo.state != process.BHProcessed || bfd.hasEarlierSameHashWithState(hdrInfos, index) {
			continue
		}
		if processedHash != nil {
			return false
		}

		processedHash = hdrInfo.hash
	}

	return bytes.Equal(processedHash, selectedHash)
}

func (bfd *baseForkDetector) isCompleteLosingSuffix(firstNonce uint64, rawProbable uint64) bool {
	previousHdrInfos := bfd.headers[firstNonce]
	if firstNonce == 0 {
		return false
	}
	if firstNonce == rawProbable {
		return true
	}

	for nonce := firstNonce + 1; ; nonce++ {
		hdrInfos := bfd.headers[nonce]
		if !bfd.hasProvenHeader(hdrInfos) || !bfd.allProvenHeadersHaveKnownV3Ancestry(hdrInfos) {
			return false
		}
		if !bfd.allProvenHeadersExtendFrontier(hdrInfos, previousHdrInfos) {
			return false
		}

		if nonce == rawProbable {
			return true
		}

		previousHdrInfos = hdrInfos
	}
}

func (bfd *baseForkDetector) hasProvenHeader(hdrInfos []*headerInfo) bool {
	for _, hdrInfo := range hdrInfos {
		if hdrInfo.hasProof {
			return true
		}
	}

	return false
}

func (bfd *baseForkDetector) allProvenHeadersHaveKnownV3Ancestry(hdrInfos []*headerInfo) bool {
	for index, hdrInfo := range hdrInfos {
		if !hdrInfo.hasProof || bfd.hasEarlierSameHash(hdrInfos, index) {
			continue
		}
		if len(hdrInfo.prevHash) == 0 || !bfd.isAsyncExecutionEnabled(hdrInfo) {
			return false
		}
	}

	return true
}

func (bfd *baseForkDetector) allProvenHeadersExtendFrontier(hdrInfos []*headerInfo, previousHdrInfos []*headerInfo) bool {
	for index, hdrInfo := range hdrInfos {
		if !hdrInfo.hasProof || bfd.hasEarlierSameHash(hdrInfos, index) {
			continue
		}
		if !bfd.hasProvenHash(previousHdrInfos, hdrInfo.prevHash) {
			return false
		}
	}

	return true
}

func (bfd *baseForkDetector) hasEarlierSameHash(hdrInfos []*headerInfo, index int) bool {
	for previousIndex := 0; previousIndex < index; previousIndex++ {
		if bytes.Equal(hdrInfos[previousIndex].hash, hdrInfos[index].hash) {
			return true
		}
	}

	return false
}

func (bfd *baseForkDetector) hasEarlierSameHashWithState(hdrInfos []*headerInfo, index int) bool {
	for previousIndex := 0; previousIndex < index; previousIndex++ {
		if hdrInfos[previousIndex].state == hdrInfos[index].state &&
			bytes.Equal(hdrInfos[previousIndex].hash, hdrInfos[index].hash) {
			return true
		}
	}

	return false
}

func (bfd *baseForkDetector) hasProvenHash(hdrInfos []*headerInfo, hash []byte) bool {
	for _, hdrInfo := range hdrInfos {
		if hdrInfo.hasProof && bytes.Equal(hdrInfo.hash, hash) {
			return true
		}
	}

	return false
}

func (bfd *baseForkDetector) isAsyncExecutionEnabled(hdrInfo *headerInfo) bool {
	return common.IsAsyncExecutionEnabledForEpochAndRound(
		bfd.enableEpochsHandler,
		bfd.enableRoundsHandler,
		hdrInfo.epoch,
		hdrInfo.round,
	)
}

// RemoveHeader removes the stored header with the given nonce and hash
func (bfd *baseForkDetector) RemoveHeader(nonce uint64, hash []byte) {
	finalCheckpointNonce := bfd.finalCheckpoint().nonce
	if nonce <= finalCheckpointNonce {
		log.Debug("baseForkDetector.RemoveHeader: given nonce is lower or equal than final checkpoint",
			"nonce", nonce,
			"final checkpoint nonce", finalCheckpointNonce)
		return
	}

	if bfd.proofsPool.HasProof(bfd.shardID, hash) {
		log.Debug("baseForkDetector.RemoveHeader: proof available for the given header, skipping removal",
			"nonce", nonce,
			"hash", hash,
			"final checkpoint nonce", finalCheckpointNonce)
		return
	}

	bfd.removeCheckpointWithNonce(nonce)

	preservedHdrsInfo := make([]*headerInfo, 0)

	bfd.mutHeaders.Lock()

	hdrsInfo := bfd.headers[nonce]
	for _, hdrInfo := range hdrsInfo {
		if hdrInfo.state != process.BHNotarized && bytes.Equal(hash, hdrInfo.hash) {
			continue
		}

		preservedHdrsInfo = append(preservedHdrsInfo, hdrInfo)
	}

	if len(preservedHdrsInfo) == 0 {
		delete(bfd.headers, nonce)
	} else {
		bfd.headers[nonce] = preservedHdrsInfo
	}

	bfd.mutHeaders.Unlock()

	bfd.forkDetector.computeFinalCheckpoint()

	probableHighestNonce := bfd.recomputeProbableHighestNonce()

	log.Debug("forkDetector.RemoveHeader",
		"nonce", nonce,
		"hash", hash,
		"probable highest nonce", probableHighestNonce,
		"final checkpoint nonce", bfd.finalCheckpoint().nonce)
}

// RemoveCommittedHeader removes a reverted committed header together with its checkpoint, proof
// included, so a same-nonce sibling can be adopted; it never removes at or below the final checkpoint
func (bfd *baseForkDetector) RemoveCommittedHeader(nonce uint64, hash []byte) {
	finalCheckpointNonce := bfd.finalCheckpoint().nonce
	if nonce <= finalCheckpointNonce {
		log.Warn("baseForkDetector.RemoveCommittedHeader: refusing removal at or below the final checkpoint",
			"nonce", nonce,
			"hash", hash,
			"final checkpoint nonce", finalCheckpointNonce)
		return
	}

	bfd.removeCheckpointWithNonce(nonce)

	preservedHdrsInfo := make([]*headerInfo, 0)

	bfd.mutHeaders.Lock()

	hdrsInfo := bfd.headers[nonce]
	for _, hdrInfo := range hdrsInfo {
		if hdrInfo.state != process.BHNotarized && bytes.Equal(hash, hdrInfo.hash) {
			continue
		}

		preservedHdrsInfo = append(preservedHdrsInfo, hdrInfo)
	}

	if len(preservedHdrsInfo) == 0 {
		delete(bfd.headers, nonce)
	} else {
		bfd.headers[nonce] = preservedHdrsInfo
	}

	bfd.mutHeaders.Unlock()

	bfd.forkDetector.computeFinalCheckpoint()

	probableHighestNonce := bfd.recomputeProbableHighestNonce()

	log.Debug("forkDetector.RemoveCommittedHeader",
		"nonce", nonce,
		"hash", hash,
		"probable highest nonce", probableHighestNonce,
		"final checkpoint nonce", bfd.finalCheckpoint().nonce)
}

// ReconcileFinalCheckpoint lowers the final checkpoint below the given nonce; this is the only
// sanctioned finality regression, gated on proven equivocation evidence (the reconcile backstop)
func (bfd *baseForkDetector) ReconcileFinalCheckpoint(nonce uint64) {
	if nonce == 0 {
		return
	}
	// only the exact final nonce may be reconciled: a higher final means settled descendants exist
	if bfd.finalCheckpoint().nonce != nonce {
		return
	}

	newFinal := &checkpointInfo{nonce: nonce - 1}

	bfd.mutFork.Lock()
	for _, checkpoint := range bfd.fork.checkpoint {
		if checkpoint.nonce < nonce && checkpoint.nonce >= newFinal.nonce {
			newFinal = checkpoint
		}
	}
	bfd.fork.finalCheckpoint = newFinal
	bfd.mutFork.Unlock()

	log.Error("forkDetector.ReconcileFinalCheckpoint: final checkpoint lowered on equivocation evidence",
		"nonce", nonce,
		"new final nonce", newFinal.nonce,
		"new final hash", newFinal.hash)
}

// ReconcileFinalCheckpointBelow purges every record and checkpoint at or above the nonce, records
// first so no concurrent recomputation re-advances from a purged entry, then lowers the final
// checkpoint below it; refused at or under the forward-only settled checkpoint
func (bfd *baseForkDetector) ReconcileFinalCheckpointBelow(nonce uint64) bool {
	if nonce == 0 {
		return false
	}

	settledNonce := bfd.settledCheckpoint().nonce
	if nonce <= settledNonce {
		log.Error("forkDetector.ReconcileFinalCheckpointBelow: refused, would cross the settled checkpoint",
			"nonce", nonce,
			"settled checkpoint nonce", settledNonce)
		return false
	}

	bfd.mutHeaders.Lock()
	for hdrNonce := range bfd.headers {
		if hdrNonce >= nonce {
			delete(bfd.headers, hdrNonce)
		}
	}
	bfd.mutHeaders.Unlock()

	bfd.mutFork.Lock()
	newFinal := &checkpointInfo{nonce: nonce - 1}
	preservedCheckpoints := make([]*checkpointInfo, 0, len(bfd.fork.checkpoint))
	for _, checkpoint := range bfd.fork.checkpoint {
		if checkpoint.nonce >= nonce {
			continue
		}

		preservedCheckpoints = append(preservedCheckpoints, checkpoint)
		if checkpoint.nonce >= newFinal.nonce {
			newFinal = checkpoint
		}
	}
	bfd.fork.checkpoint = preservedCheckpoints
	loweredFinal := bfd.fork.finalCheckpoint.nonce >= nonce
	if loweredFinal {
		bfd.fork.finalCheckpoint = newFinal
	}
	bfd.mutFork.Unlock()

	bfd.recomputeProbableHighestNonce()

	log.Error("forkDetector.ReconcileFinalCheckpointBelow: finality regressed on dead cross-notarization evidence",
		"nonce", nonce,
		"final checkpoint lowered", loweredFinal,
		"new final nonce", bfd.finalCheckpoint().nonce)

	return true
}

func (bfd *baseForkDetector) removeCheckpointWithNonce(nonce uint64) {
	bfd.mutFork.Lock()
	preservedCheckpoint := make([]*checkpointInfo, 0)

	for i := 0; i < len(bfd.fork.checkpoint); i++ {
		if bfd.fork.checkpoint[i].nonce == nonce {
			continue
		}

		preservedCheckpoint = append(preservedCheckpoint, bfd.fork.checkpoint[i])
	}

	bfd.fork.checkpoint = preservedCheckpoint
	bfd.mutFork.Unlock()

	log.Debug("forkDetector.removeCheckpointWithNonce",
		"nonce", nonce,
		"last checkpoint nonce", bfd.lastCheckpoint().nonce)
}

// append adds a new header in the slice found in nonce position
// it not adds the header if its hash is already stored in the slice
func (bfd *baseForkDetector) append(hdrInfo *headerInfo) bool {
	return bfd.appendHeaderInfo(hdrInfo).inserted
}

func (bfd *baseForkDetector) appendHeaderInfo(hdrInfo *headerInfo) appendHeaderInfoResult {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	hdrInfos := bfd.headers[hdrInfo.nonce]
	isHdrInfosNilOrEmpty := len(hdrInfos) == 0 // no need for nil check, len() for nil returns 0
	if isHdrInfosNilOrEmpty {
		bfd.headers[hdrInfo.nonce] = []*headerInfo{hdrInfo}
		return appendHeaderInfoResult{inserted: true}
	}

	enriched := bfd.adjustHeadersWithInfo(hdrInfo)

	for _, hdrInfoStored := range hdrInfos {
		if bytes.Equal(hdrInfoStored.hash, hdrInfo.hash) && hdrInfoStored.state == hdrInfo.state && hdrInfoStored.hasProof == hdrInfo.hasProof {
			return appendHeaderInfoResult{enriched: enriched}
		}
	}

	bfd.headers[hdrInfo.nonce] = append(bfd.headers[hdrInfo.nonce], hdrInfo)
	return appendHeaderInfoResult{inserted: true, enriched: enriched}
}

func (bfd *baseForkDetector) adjustHeadersWithInfo(hInfo *headerInfo) bool {
	enriched := false
	canEnrichAncestry := len(hInfo.prevHash) > 0 && bfd.isAsyncExecutionEnabled(hInfo)

	hdrInfos := bfd.headers[hInfo.nonce]
	for i := range hdrInfos {
		if !bytes.Equal(hdrInfos[i].hash, hInfo.hash) {
			continue
		}

		if hInfo.hasProof && !hdrInfos[i].hasProof && bfd.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, hInfo.epoch) {
			hdrInfos[i].hasProof = true
			enriched = true
		}
		if canEnrichAncestry && len(hdrInfos[i].prevHash) == 0 {
			hdrInfos[i].prevHash = hInfo.prevHash
			enriched = true
		}
	}

	return enriched
}

// GetHighestFinalBlockNonce gets the highest nonce of the block which is final, and it can not be reverted anymore
func (bfd *baseForkDetector) GetHighestFinalBlockNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

// GetHighestFinalBlockHash gets the hash of the block which is final, and it can not be reverted anymore
func (bfd *baseForkDetector) GetHighestFinalBlockHash() []byte {
	return bfd.finalCheckpoint().hash
}

// ProbableHighestNonce gets the probable highest nonce
func (bfd *baseForkDetector) ProbableHighestNonce() uint64 {
	return bfd.probableHighestNonce()
}

// ResetFork resets the forced fork
func (bfd *baseForkDetector) ResetFork() {
	bfd.ResetProbableHighestNonce()
	bfd.setLastRoundWithForcedFork(bfd.roundHandler.Index())

	log.Debug("forkDetector.ResetFork",
		"last round with forced fork", bfd.lastRoundWithForcedFork())
}

// ResetProbableHighestNonce resets the probable highest nonce to the last checkpoint nonce / the highest notarized nonce
func (bfd *baseForkDetector) ResetProbableHighestNonce() {
	bfd.cleanupReceivedHeadersHigherThanNonce(bfd.lastCheckpoint().nonce)
	bfd.recomputeProbableHighestNonce()

	log.Debug("forkDetector.ResetProbableHighestNonce",
		"probable highest nonce", bfd.probableHighestNonce())
}

func (bfd *baseForkDetector) addCheckpoint(checkpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	bfd.fork.checkpoint = append(bfd.fork.checkpoint, checkpoint)
	bfd.mutFork.Unlock()
}

// AddCheckpoint adds a new checkpoint in the list
func (bfd *baseForkDetector) AddCheckpoint(nonce uint64, round uint64, hash []byte) {
	checkpoint := &checkpointInfo{
		nonce: nonce,
		round: round,
		hash:  hash,
	}
	bfd.addCheckpoint(checkpoint)
}

func (bfd *baseForkDetector) lastCheckpoint() *checkpointInfo {
	bfd.mutFork.RLock()
	lastIndex := len(bfd.fork.checkpoint) - 1
	if lastIndex < 0 {
		bfd.mutFork.RUnlock()
		return &checkpointInfo{
			nonce: bfd.genesisNonce,
			round: bfd.genesisRound,
		}
	}
	lastCheckpoint := bfd.fork.checkpoint[lastIndex]
	bfd.mutFork.RUnlock()

	return lastCheckpoint
}

func (bfd *baseForkDetector) setFinalCheckpoint(finalCheckpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	bfd.fork.finalCheckpoint = finalCheckpoint
	bfd.mutFork.Unlock()
}

// advanceFinalCheckpoint sets the final checkpoint only forward, so concurrent computations
// cannot regress an already finalized nonce
func (bfd *baseForkDetector) advanceFinalCheckpoint(finalCheckpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	if finalCheckpoint.nonce > bfd.fork.finalCheckpoint.nonce {
		bfd.fork.finalCheckpoint = finalCheckpoint
	}
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) setSettledCheckpoint(settledCheckpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	bfd.fork.settledCheckpoint = settledCheckpoint
	bfd.mutFork.Unlock()
}

// advanceSettledCheckpoint sets the settled checkpoint only forward; settlement is never undone
func (bfd *baseForkDetector) advanceSettledCheckpoint(settledCheckpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	if settledCheckpoint.nonce > bfd.fork.settledCheckpoint.nonce {
		bfd.fork.settledCheckpoint = settledCheckpoint
	}
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) settledCheckpoint() *checkpointInfo {
	bfd.mutFork.RLock()
	settledCheckpoint := bfd.fork.settledCheckpoint
	bfd.mutFork.RUnlock()

	return settledCheckpoint
}

// GetHighestSettledBlockInfo gets the nonce and hash of the settled block as a consistent pair;
// unlike the final checkpoint, the settled one is settlement-anchored and never reconciled
func (bfd *baseForkDetector) GetHighestSettledBlockInfo() (uint64, []byte) {
	settledCheckpoint := bfd.settledCheckpoint()

	return settledCheckpoint.nonce, settledCheckpoint.hash
}

func (bfd *baseForkDetector) isSupernovaForHeader(header data.HeaderHandler) bool {
	return bfd.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, header.GetEpoch())
}

func isParentCheckpoint(checkpoint *checkpointInfo, header data.HeaderHandler) bool {
	if checkpoint.nonce+1 != header.GetNonce() {
		return false
	}

	return len(checkpoint.hash) == 0 || bytes.Equal(checkpoint.hash, header.GetPrevHash())
}

// canInstantlyFinalize returns false when settlement evidence is still required.
func (bfd *baseForkDetector) canInstantlyFinalize(header data.HeaderHandler, headerHash []byte) bool {
	if !bfd.isSupernovaForHeader(header) {
		return true
	}

	finalCheckpoint := bfd.finalCheckpoint()
	if !isParentCheckpoint(finalCheckpoint, header) {
		return false
	}
	if common.IsCrossHeaderSettlementEnabledForHeader(bfd.enableEpochsHandler, bfd.enableRoundsHandler, header) &&
		bfd.hasCompetingSiblingEvidence(header.GetNonce(), headerHash, header.GetPrevHash()) {
		return false
	}

	return !common.IsContendedRound(header.GetRound(), finalCheckpoint.round)
}

func (bfd *baseForkDetector) hasCompetingSiblingEvidence(nonce uint64, hash []byte, parentHash []byte) bool {
	bfd.mutHeaders.RLock()
	defer bfd.mutHeaders.RUnlock()

	return bfd.hasCompetingSiblingEvidenceLocked(nonce, hash, parentHash)
}

func (bfd *baseForkDetector) hasCompetingSiblingEvidenceLocked(nonce uint64, hash []byte, parentHash []byte) bool {
	for _, hdrInfo := range bfd.headers[nonce] {
		if (hdrInfo.hasProof || hdrInfo.state == process.BHNotarized) &&
			!bytes.Equal(hdrInfo.hash, hash) &&
			bytes.Equal(hdrInfo.prevHash, parentHash) &&
			bfd.isAsyncExecutionEnabled(hdrInfo) {
			return true
		}
	}

	return false
}

// RestoreToGenesis sets class variables to theirs initial values
func (bfd *baseForkDetector) RestoreToGenesis() {
	bfd.mutProbableHighestNonceUpdate.Lock()
	defer bfd.mutProbableHighestNonceUpdate.Unlock()

	bfd.mutHeaders.Lock()
	bfd.headers = make(map[uint64][]*headerInfo)
	bfd.mutHeaders.Unlock()

	bfd.mutFork.Lock()

	checkpoint := &checkpointInfo{
		nonce: bfd.genesisNonce,
		round: bfd.genesisRound,
	}
	bfd.fork.checkpoint = []*checkpointInfo{checkpoint}
	bfd.fork.finalCheckpoint = checkpoint
	bfd.fork.settledCheckpoint = checkpoint
	bfd.fork.probableHighestNonce = bfd.genesisNonce
	bfd.fork.highestNonceReceived = bfd.genesisNonce
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) finalCheckpoint() *checkpointInfo {
	bfd.mutFork.RLock()
	finalCheckpoint := bfd.fork.finalCheckpoint
	bfd.mutFork.RUnlock()

	return finalCheckpoint
}

func (bfd *baseForkDetector) setProbableHighestNonce(nonce uint64) {
	bfd.mutFork.Lock()
	if bfd.shardID != core.MetachainShardId && nonce < bfd.fork.finalCheckpoint.nonce {
		nonce = bfd.fork.finalCheckpoint.nonce
	}
	bfd.fork.probableHighestNonce = nonce
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) recomputeProbableHighestNonce() uint64 {
	bfd.mutProbableHighestNonceUpdate.Lock()
	defer bfd.mutProbableHighestNonceUpdate.Unlock()

	probableHighestNonce := bfd.computeProbableHighestNonce()
	bfd.setProbableHighestNonce(probableHighestNonce)

	return probableHighestNonce
}

func (bfd *baseForkDetector) probableHighestNonce() uint64 {
	bfd.mutFork.RLock()
	probableHighestNonce := bfd.fork.probableHighestNonce
	bfd.mutFork.RUnlock()

	return probableHighestNonce
}

func (bfd *baseForkDetector) setHighestNonceReceived(nonce uint64) {
	if nonce <= bfd.highestNonceReceived() {
		return
	}

	bfd.mutFork.Lock()
	bfd.fork.highestNonceReceived = nonce
	bfd.mutFork.Unlock()

	log.Debug("forkDetector.setHighestNonceReceived",
		"highest nonce received", nonce)
}

func (bfd *baseForkDetector) highestNonceReceived() uint64 {
	bfd.mutFork.RLock()
	highestNonceReceived := bfd.fork.highestNonceReceived
	bfd.mutFork.RUnlock()

	return highestNonceReceived
}

// logFinalityLag exposes how far the final checkpoint trails the received frontier;
// a steadily growing lag means finality stopped advancing while the chain moved on
func (bfd *baseForkDetector) logFinalityLag() {
	finalNonce := bfd.finalCheckpoint().nonce
	highestNonce := bfd.highestNonceReceived()
	lag := uint64(0)
	if highestNonce > finalNonce {
		lag = highestNonce - finalNonce
	}

	log.Debug("forkDetector finality lag",
		"final checkpoint nonce", finalNonce,
		"settled checkpoint nonce", bfd.settledCheckpoint().nonce,
		"probable highest nonce", bfd.probableHighestNonce(),
		"highest received nonce", highestNonce,
		"lag", lag,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bfd *baseForkDetector) IsInterfaceNil() bool {
	return bfd == nil
}

// CheckFork method checks if the node could be on the fork
func (bfd *baseForkDetector) CheckFork() *process.ForkInfo {
	var (
		forkHeaderRound uint64
		forkHeaderHash  []byte
		selfHdrInfo     *headerInfo
		forkHeaderEpoch uint32
	)

	forkInfoObject := process.NewForkInfo()

	if bfd.isConsensusStuck() {
		forkInfoObject.IsDetected = true
		return forkInfoObject
	}

	rollBackNonce := bfd.getRollBackNonce()
	if rollBackNonce < math.MaxUint64 {
		forkInfoObject.IsDetected = true
		forkInfoObject.Nonce = rollBackNonce
		bfd.SetRollBackNonce(math.MaxUint64)
		return forkInfoObject
	}

	finalCheckpointNonce := bfd.finalCheckpoint().nonce

	bfd.mutHeaders.Lock()
	for nonce, hdrsInfo := range bfd.headers {
		if len(hdrsInfo) == 1 {
			continue
		}
		if nonce <= finalCheckpointNonce {
			continue
		}

		selfHdrInfo = getProcessedHeaderInfo(hdrsInfo)
		if selfHdrInfo == nil {
			continue
		}

		forkHeaderRound = math.MaxUint64
		forkHeaderHash = nil
		forkHeaderEpoch = 0
		bfd.maxForkHeaderEpoch = selfHdrInfo.epoch
		for _, hdrInfo := range hdrsInfo {
			if hdrInfo.state == process.BHProcessed ||
				!bfd.isForkCandidateForProcessedHeader(selfHdrInfo, hdrInfo) {
				continue
			}
			if hdrInfo.epoch > bfd.maxForkHeaderEpoch {
				bfd.maxForkHeaderEpoch = hdrInfo.epoch
			}
		}

		for i := 0; i < len(hdrsInfo); i++ {
			if hdrsInfo[i].state == process.BHProcessed {
				continue
			}
			if !bfd.isForkCandidateForProcessedHeader(selfHdrInfo, hdrsInfo[i]) {
				continue
			}

			forkHeaderHash, forkHeaderRound, forkHeaderEpoch = bfd.computeForkInfo(
				hdrsInfo[i],
				forkHeaderHash,
				forkHeaderRound,
				forkHeaderEpoch,
			)
		}

		if bfd.shouldSignalFork(selfHdrInfo, forkHeaderHash, forkHeaderRound, forkHeaderEpoch) {
			forkInfoObject.IsDetected = true
			if nonce < forkInfoObject.Nonce {
				forkInfoObject.Nonce = nonce
				forkInfoObject.Round = forkHeaderRound
				forkInfoObject.Hash = forkHeaderHash
			}
		}
	}
	bfd.mutHeaders.Unlock()

	return forkInfoObject
}

func getProcessedHeaderInfo(hdrInfos []*headerInfo) *headerInfo {
	var processedHeader *headerInfo
	for _, hdrInfo := range hdrInfos {
		if hdrInfo.state == process.BHProcessed {
			processedHeader = hdrInfo
		}
	}

	return processedHeader
}

func (bfd *baseForkDetector) isForkCandidateForProcessedHeader(processedHeader *headerInfo, candidate *headerInfo) bool {
	if !bfd.isAsyncExecutionEnabled(processedHeader) ||
		!bfd.isAsyncExecutionEnabled(candidate) ||
		candidate.state == process.BHNotarized ||
		len(processedHeader.prevHash) == 0 {
		return true
	}

	return len(candidate.prevHash) > 0 && bytes.Equal(candidate.prevHash, processedHeader.prevHash)
}

func (bfd *baseForkDetector) computeForkInfo(
	hdrInfo *headerInfo,
	lastForkHash []byte,
	lastForkRound uint64,
	lastForkEpoch uint32,
) ([]byte, uint64, uint32) {

	if hdrInfo.state == process.BHReceivedTooLate && bfd.highestNonceReceived() > hdrInfo.nonce {
		return lastForkHash, lastForkRound, lastForkEpoch
	}

	currentForkRound := hdrInfo.round
	if hdrInfo.state == process.BHNotarized {
		currentForkRound = process.MinForkRound
	} else {
		if hdrInfo.epoch < bfd.maxForkHeaderEpoch {
			return lastForkHash, lastForkRound, lastForkEpoch
		}
	}

	if isLowerRoundOrHash(currentForkRound, hdrInfo.hash, lastForkRound, lastForkHash) {
		return hdrInfo.hash, currentForkRound, hdrInfo.epoch
	}

	return lastForkHash, lastForkRound, lastForkEpoch
}

func isLowerRoundOrHash(round uint64, hash []byte, otherRound uint64, otherHash []byte) bool {
	return round < otherRound || round == otherRound && bytes.Compare(hash, otherHash) < 0
}

func (bfd *baseForkDetector) shouldSignalFork(
	headerInfo *headerInfo,
	lastForkHash []byte,
	lastForkRound uint64,
	lastForkEpoch uint32,
) bool {
	sameHash := bytes.Equal(headerInfo.hash, lastForkHash)
	if sameHash {
		return false
	}

	if lastForkRound != process.MinForkRound {
		if headerInfo.epoch > lastForkEpoch {
			log.Trace("shouldSignalFork epoch change false")
			return false
		}

		if headerInfo.epoch < lastForkEpoch {
			log.Trace("shouldSignalFork epoch change true")
			return true
		}
	}

	higherHashForSameRound := headerInfo.round == lastForkRound &&
		bytes.Compare(headerInfo.hash, lastForkHash) > 0
	higherNonceReceived := bfd.highestNonceReceived() > headerInfo.nonce
	shouldSignalFork := headerInfo.round > lastForkRound || (higherHashForSameRound && !higherNonceReceived)

	return shouldSignalFork
}

func (bfd *baseForkDetector) isHeaderReceivedTooLate(
	header data.HeaderHandler,
	state process.BlockHeaderState,
	finality int64,
) bool {
	if state == process.BHProcessed {
		return false
	}

	// This condition would avoid a stuck situation, when shards would set as final, block with nonce n received from
	// meta-chain, because they also received n+1. In the same time meta-chain would be reverted to an older block with
	// nonce n received it with latency but before n+1. Actually this condition would reject these older blocks.
	isHeaderReceivedTooLate := int64(header.GetRound()) < bfd.roundHandler.Index()-finality

	return isHeaderReceivedTooLate
}

func (bfd *baseForkDetector) isConsensusStuck() bool {
	if bfd.lastRoundWithForcedFork() == bfd.roundHandler.Index() {
		return false
	}

	if bfd.isSyncing() {
		return false
	}

	lastCheckpoint := bfd.lastCheckpoint()
	roundsDifference := bfd.roundHandler.Index() - int64(lastCheckpoint.round)
	if roundsDifference <= bfd.getMaxRoundsWithoutCommittedBlock(uint64(bfd.roundHandler.Index())) {
		return false
	}

	if !process.IsInProperRound(bfd.roundHandler.Index()) {
		return false
	}

	// never blind-rollback a proven block: a proven tip can only be wrong through equivocation,
	// which the evidence-driven rollback paths detect and prove before acting
	hasProvenTip := len(lastCheckpoint.hash) != 0 && bfd.proofsPool.HasProof(bfd.shardID, lastCheckpoint.hash)

	return !hasProvenTip
}

func (bfd *baseForkDetector) getMaxRoundsWithoutCommittedBlock(round uint64) int64 {
	return int64(bfd.processConfigsHandler.GetMaxRoundsWithoutCommittedBlock(round))
}

func (bfd *baseForkDetector) isSyncing() bool {
	noncesDifference := int64(bfd.ProbableHighestNonce()) - int64(bfd.lastCheckpoint().nonce)
	isSyncing := noncesDifference > process.NonceDifferenceWhenSynced
	return isSyncing
}

// GetNotarizedHeaderHash returns the notarized header hash at nonce.
func (bfd *baseForkDetector) GetNotarizedHeaderHash(nonce uint64) []byte {
	hash, _, _ := bfd.getNotarizedHeaderSelection(nonce)

	return hash
}

func (bfd *baseForkDetector) getNotarizedHeaderSelection(nonce uint64) ([]byte, bool, bool) {
	bfd.mutHeaders.RLock()
	defer bfd.mutHeaders.RUnlock()

	hdrInfos := bfd.headers[nonce]
	var selectedHeader *headerInfo
	numHashes := 0
	hasV3Header := false
	for index, hdrInfo := range hdrInfos {
		if hdrInfo.state != process.BHNotarized || bfd.hasEarlierSameHashWithState(hdrInfos, index) {
			continue
		}

		if selectedHeader == nil {
			selectedHeader = hdrInfo
		}
		numHashes++
		hasV3Header = hasV3Header || bfd.isAsyncExecutionEnabled(hdrInfo)
	}

	if numHashes > 1 && hasV3Header {
		return nil, false, true
	}
	if selectedHeader != nil {
		return selectedHeader.hash, bfd.isAsyncExecutionEnabled(selectedHeader), false
	}

	return nil, false, false
}

func (bfd *baseForkDetector) getHeaderVersion(nonce uint64, hash []byte) (bool, bool) {
	bfd.mutHeaders.RLock()
	defer bfd.mutHeaders.RUnlock()

	for _, hdrInfo := range bfd.headers[nonce] {
		if bytes.Equal(hdrInfo.hash, hash) {
			return bfd.isAsyncExecutionEnabled(hdrInfo), true
		}
	}

	return false, false
}

func (bfd *baseForkDetector) cleanupReceivedHeadersHigherThanNonce(nonce uint64) {
	bfd.mutHeaders.Lock()
	for hdrsNonce, hdrsInfo := range bfd.headers {
		if hdrsNonce <= nonce {
			continue
		}

		preservedHdrsInfo := make([]*headerInfo, 0)

		for _, hdrInfo := range hdrsInfo {
			// a proven record is hard evidence of the network tip; purging it would let the probable
			// nonce collapse below a proven block and re-arm same-nonce proposals
			isProvenRecord := hdrInfo.hasProof &&
				bfd.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, hdrInfo.epoch)
			if hdrInfo.state != process.BHNotarized && !isProvenRecord {
				continue
			}

			preservedHdrsInfo = append(preservedHdrsInfo, hdrInfo)
		}

		if len(preservedHdrsInfo) == 0 {
			delete(bfd.headers, hdrsNonce)
			continue
		}

		bfd.headers[hdrsNonce] = preservedHdrsInfo
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *baseForkDetector) checkGenesisTimeForHeaderBeforeSupernova(
	headerHandler data.HeaderHandler,
) error {
	chainParams, err := bfd.chainParametersHandler.ChainParametersForEpoch(headerHandler.GetEpoch())
	if err != nil {
		return err
	}
	roundDuration := int64(chainParams.RoundDuration)

	// The round duration is provided as milliseconds in the configuration. It needs to be
	// converted to seconds to ensure correct calculations for genesis time before
	// supernova activation.
	roundDuration /= 1000

	roundDifference := int64(headerHandler.GetRound() - bfd.genesisRound)
	genesisTime := int64(headerHandler.GetTimeStamp()) - roundDifference*roundDuration

	if genesisTime != bfd.genesisTime {
		log.Error("checkGenesisTimeForHeaderBeforeSupernova: genesis time mismatch",
			"localGenesisTime", bfd.genesisTime,
			"calculatedGenesisTime", genesisTime,
			"header timestamp", headerHandler.GetTimeStamp(),
		)

		return ErrGenesisTimeMissmatch
	}

	return nil
}

func (bfd *baseForkDetector) getPrevSupernovaActivationEpoch(currentEpoch uint32) uint32 {
	// in this interval, chain parameters have to be taken from the epoch previous to supernova
	if currentEpoch == 0 {
		return currentEpoch
	}

	return currentEpoch - 1
}

func (bfd *baseForkDetector) checkGenesisTimeForHeaderAfterSupernovaWithoutRoundActivation(
	headerHandler data.HeaderHandler,
) error {
	chainParams, err := bfd.chainParametersHandler.ChainParametersForEpoch(bfd.getPrevSupernovaActivationEpoch(headerHandler.GetEpoch()))
	if err != nil {
		return err
	}
	roundDuration := int64(chainParams.RoundDuration)
	roundDifference := int64(headerHandler.GetRound() - bfd.genesisRound)
	genesisTime := int64(headerHandler.GetTimeStamp()) - roundDifference*roundDuration

	log.Trace("getGenesisTimeForHeaderAfterSupernovaWithoutRoundActivation",
		"roundDuration", roundDuration,
		"roundDifference", roundDifference,
		"calculated genesisTime", genesisTime,
		"genesisTime", bfd.genesisTime,
	)

	// if supernova is activated from genesis (epoch zero) this reduction is not needed since
	// genesisTime from config will be directly as milliseconds; otherwise it has to be
	// reduced to seconds granularity, in this specific interval (when supernova epoch is
	// activated but supernova round is not yet activated)
	supernovaActivatedInEpochZero := bfd.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, 0)
	if !supernovaActivatedInEpochZero {
		genesisTime /= 1000
	}

	if genesisTime != bfd.genesisTime {
		log.Error("checkGenesisTimeForHeaderAfterSupernovaWithoutRoundActivation: genesis time mismatch",
			"localGenesisTime", bfd.genesisTime,
			"calculatedGenesisTime", genesisTime,
			"header timestamp", headerHandler.GetTimeStamp(),
		)
		return ErrGenesisTimeMissmatch
	}

	return nil
}

func (bfd *baseForkDetector) checkGenesisTimeForHeaderAfterSupernovaWithRoundActivation(
	headerHandler data.HeaderHandler,
) error {
	activationRound := bfd.enableRoundsHandler.GetActivationRound(common.SupernovaRoundFlag)

	chainParams, err := bfd.chainParametersHandler.ChainParametersForEpoch(headerHandler.GetEpoch())
	if err != nil {
		return err
	}
	roundDuration := int64(chainParams.RoundDuration)

	roundDifference := int64(headerHandler.GetRound()) - int64(activationRound)
	if roundDifference < 0 {
		log.Warn("current round lower than supernova activation round",
			"current round", headerHandler.GetRound(),
			"supernova activationRound", activationRound,
		)

		return ErrGenesisTimeMissmatch
	}

	genesisTime := int64(headerHandler.GetTimeStamp()) - roundDifference*roundDuration

	log.Trace("getGenesisTimeForHeaderAfterSupernovaWithRoundActivation",
		"activationRound", activationRound,
		"roundDuration", roundDuration,
		"roundDifference", roundDifference,
		"genesisTime", genesisTime,
		"supernovaGenesisTime", bfd.supernovaGenesisTime,
	)

	if genesisTime != bfd.supernovaGenesisTime {
		log.Error("checkGenesisTimeForHeaderAfterSupernovaWithRoundActivation: genesis time mismatch",
			"localGenesisTime", bfd.supernovaGenesisTime,
			"calculatedGenesisTime", genesisTime,
			"header timestamp", headerHandler.GetTimeStamp(),
		)
		return ErrGenesisTimeMissmatch
	}

	return nil
}

func (bfd *baseForkDetector) checkGenesisTimeForHeader(headerHandler data.HeaderHandler) error {
	supernovaInEpochActivated := bfd.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, headerHandler.GetEpoch())
	supernovaInRoundActivated := bfd.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, headerHandler.GetRound())

	if !supernovaInEpochActivated {
		return bfd.checkGenesisTimeForHeaderBeforeSupernova(headerHandler)
	}

	if !supernovaInRoundActivated {
		return bfd.checkGenesisTimeForHeaderAfterSupernovaWithoutRoundActivation(headerHandler)
	}

	return bfd.checkGenesisTimeForHeaderAfterSupernovaWithRoundActivation(headerHandler)
}

func (bfd *baseForkDetector) addHeader(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
	doJobOnBHProcessed func(data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte),
) error {

	err := bfd.checkBlockBasicValidity(header, headerHash)
	if err != nil {
		return err
	}

	bfd.processReceivedBlock(header, headerHash, state, selfNotarizedHeaders, selfNotarizedHeadersHashes, doJobOnBHProcessed)
	return nil
}

// ReceivedProof is called when a proof is received
func (bfd *baseForkDetector) ReceivedProof(proof data.HeaderProofHandler) {
	bfd.processReceivedProof(proof)
}

func (bfd *baseForkDetector) processReceivedProof(proof data.HeaderProofHandler) {
	bfd.setHighestNonceReceived(proof.GetHeaderNonce())

	hInfo := &headerInfo{
		epoch:    proof.GetHeaderEpoch(),
		nonce:    proof.GetHeaderNonce(),
		round:    proof.GetHeaderRound(),
		hash:     proof.GetHeaderHash(),
		state:    process.BHReceived,
		hasProof: true,
	}

	_ = bfd.appendHeaderInfo(hInfo)

	bfd.recomputeProbableHighestNonce()

	log.Trace("forkDetector.processReceivedProof",
		"round", hInfo.round,
		"nonce", hInfo.nonce,
		"hash", hInfo.hash,
		"state", hInfo.state,
		"probable highest nonce", bfd.probableHighestNonce(),
		"last checkpoint nonce", bfd.lastCheckpoint().nonce,
		"final checkpoint nonce", bfd.finalCheckpoint().nonce,
		"has proof", hInfo.hasProof)
}

func (bfd *baseForkDetector) processReceivedBlock(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
	doJobOnBHProcessed func(data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte),
) {
	hasProof := true // old blocks have consensus proof on them
	if common.IsProofsFlagEnabledForHeader(bfd.enableEpochsHandler, header) {
		hasProof = bfd.proofsPool.HasProof(header.GetShardID(), headerHash)
	}
	bfd.setHighestNonceReceived(header.GetNonce())

	if state == process.BHProposed || !hasProof {
		log.Trace("forkDetector.processReceivedBlock: block is proposed or has no proof", "state", state, "has proof", hasProof)
		return
	}

	isHeaderReceivedTooLate := bfd.isHeaderReceivedTooLate(header, state, process.BlockFinality)
	if isHeaderReceivedTooLate {
		log.Trace("forkDetector.processReceivedBlock: block is received too late", "initial state", state)
		state = process.BHReceivedTooLate
	}

	hInfo := &headerInfo{
		epoch:    header.GetEpoch(),
		nonce:    header.GetNonce(),
		round:    header.GetRound(),
		hash:     headerHash,
		prevHash: header.GetPrevHash(),
		state:    state,
		hasProof: hasProof,
	}

	appendResult := bfd.appendHeaderInfo(hInfo)
	if !appendResult.inserted && !appendResult.enriched {
		log.Trace("forkDetector.processReceivedBlock: header not appended", "nonce", hInfo.nonce, "hash", hInfo.hash)
		return
	}

	if appendResult.inserted && state == process.BHProcessed {
		doJobOnBHProcessed(header, headerHash, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}

	bfd.recomputeProbableHighestNonce()

	log.Debug("forkDetector.appendHeaderInfo",
		"round", hInfo.round,
		"nonce", hInfo.nonce,
		"hash", hInfo.hash,
		"state", hInfo.state,
		"probable highest nonce", bfd.probableHighestNonce(),
		"last checkpoint nonce", bfd.lastCheckpoint().nonce,
		"final checkpoint nonce", bfd.finalCheckpoint().nonce,
		"has proof", hInfo.hasProof)
}

// SetFinalToLastCheckpoint sets the final and settled checkpoints to the last checkpoint added in
// list; used only at bootstrap restore, where the persisted nonce is the settled one
func (bfd *baseForkDetector) SetFinalToLastCheckpoint() {
	lastCheckpoint := bfd.lastCheckpoint()
	bfd.setFinalCheckpoint(lastCheckpoint)
	bfd.setSettledCheckpoint(lastCheckpoint)
}
