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

var _ process.ForkDetector = (*shardForkDetector)(nil)

// shardForkDetector implements the shard fork detector mechanism
type shardForkDetector struct {
	*baseForkDetector
	mutFinalityUpdate sync.Mutex
}

// NewShardForkDetector method creates a new shardForkDetector object
func NewShardForkDetector(
	roundHandler consensus.RoundHandler,
	blackListHandler process.TimeCacher,
	blockTracker process.BlockTracker,
	genesisTime int64,
	supernovaGenesisTime int64,
	enableEpochsHandler common.EnableEpochsHandler,
	enableRoundsHandler common.EnableRoundsHandler,
	proofsPool process.ProofsPool,
	chainParametersHandler common.ChainParametersHandler,
	processConfigsHandler common.ProcessConfigsHandler,
	shardID uint32,
) (*shardForkDetector, error) {
	if check.IfNil(roundHandler) {
		return nil, process.ErrNilRoundHandler
	}
	if check.IfNil(blackListHandler) {
		return nil, process.ErrNilBlackListCacher
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(enableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}
	if check.IfNil(proofsPool) {
		return nil, process.ErrNilProofsPool
	}
	if check.IfNil(chainParametersHandler) {
		return nil, process.ErrNilChainParametersHandler
	}
	if check.IfNil(processConfigsHandler) {
		return nil, process.ErrNilProcessConfigsHandler
	}

	genesisHdr, _, err := blockTracker.GetSelfNotarizedHeader(core.MetachainShardId, 0)
	if err != nil {
		return nil, err
	}

	bfd := &baseForkDetector{
		roundHandler:           roundHandler,
		blackListHandler:       blackListHandler,
		genesisTime:            genesisTime,
		supernovaGenesisTime:   supernovaGenesisTime,
		blockTracker:           blockTracker,
		genesisNonce:           genesisHdr.GetNonce(),
		genesisRound:           genesisHdr.GetRound(),
		genesisEpoch:           genesisHdr.GetEpoch(),
		enableEpochsHandler:    enableEpochsHandler,
		enableRoundsHandler:    enableRoundsHandler,
		proofsPool:             proofsPool,
		chainParametersHandler: chainParametersHandler,
		processConfigsHandler:  processConfigsHandler,
		shardID:                shardID,
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	bfd.fork.checkpoint = make([]*checkpointInfo, 0)
	checkpoint := &checkpointInfo{
		nonce: bfd.genesisNonce,
		round: bfd.genesisRound,
	}
	bfd.setFinalAndSettledCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)
	bfd.fork.rollBackNonce = math.MaxUint64
	bfd.fork.probableHighestNonce = bfd.genesisNonce
	bfd.fork.highestNonceReceived = bfd.genesisNonce

	sfd := shardForkDetector{
		baseForkDetector: bfd,
	}

	sfd.blockTracker.RegisterSelfNotarizedFromCrossHeadersHandler(sfd.ReceivedSelfNotarizedFromCrossHeaders)

	bfd.forkDetector = &sfd

	return &sfd, nil
}

// AddHeader method adds a new header to headers map
func (sfd *shardForkDetector) AddHeader(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) error {
	return sfd.addHeader(
		header,
		headerHash,
		state,
		selfNotarizedHeaders,
		selfNotarizedHeadersHashes,
		sfd.doJobOnBHProcessed,
	)
}

func (sfd *shardForkDetector) doJobOnBHProcessed(
	header data.HeaderHandler,
	headerHash []byte,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) {
	_ = sfd.appendSelfNotarizedHeaders(selfNotarizedHeaders, selfNotarizedHeadersHashes, core.MetachainShardId)
	sfd.computeFinalCheckpoint()
	newCheckpoint := &checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash}
	sfd.addCheckpoint(newCheckpoint)
	// first shard block with proof does not have increased consensus
	// so instant finality will only be set after the first block with increased consensus
	if common.IsFlagEnabledAfterEpochsStartBlock(header, sfd.enableEpochsHandler, common.AndromedaFlag) &&
		sfd.canInstantlyFinalize(header, headerHash) {
		// under Supernova the settled checkpoint advances only on meta notarization
		if sfd.isSupernovaForHeader(header) {
			sfd.setFinalCheckpoint(newCheckpoint)
		} else {
			sfd.setFinalAndSettledCheckpoint(newCheckpoint)
		}
	}
	sfd.removePastOrInvalidRecords()
}

// ReceivedSelfNotarizedFromCrossHeaders is a registered call handler through which fork detector is notified
// when metachain notarized new headers from self shard
func (sfd *shardForkDetector) ReceivedSelfNotarizedFromCrossHeaders(
	shardID uint32,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) {
	// accept only self notarized headers by meta
	if shardID != core.MetachainShardId {
		return
	}

	appended := sfd.appendSelfNotarizedHeaders(selfNotarizedHeaders, selfNotarizedHeadersHashes, shardID)
	if appended {
		sfd.computeFinalCheckpoint()
		for _, header := range selfNotarizedHeaders {
			if header.IsHeaderV3() &&
				common.IsCrossHeaderSettlementEnabledForHeader(sfd.enableEpochsHandler, sfd.enableRoundsHandler, header) {
				sfd.recomputeProbableHighestNonce()
				break
			}
		}
	}
}

func (sfd *shardForkDetector) appendSelfNotarizedHeaders(
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
	shardID uint32,
) bool {

	selfNotarizedHeaderAdded := false
	settledNonce := sfd.settledCheckpoint().nonce

	for i := 0; i < len(selfNotarizedHeaders); i++ {
		if selfNotarizedHeaders[i].GetNonce() <= settledNonce {
			continue
		}

		hasProof := sfd.proofsPool.HasProof(selfNotarizedHeaders[i].GetShardID(), selfNotarizedHeadersHashes[i])
		appended := sfd.append(&headerInfo{
			epoch:    selfNotarizedHeaders[i].GetEpoch(),
			nonce:    selfNotarizedHeaders[i].GetNonce(),
			round:    selfNotarizedHeaders[i].GetRound(),
			hash:     selfNotarizedHeadersHashes[i],
			prevHash: selfNotarizedHeaders[i].GetPrevHash(),
			state:    process.BHNotarized,
			hasProof: hasProof,
		})
		if appended {
			log.Debug("added self notarized header in fork detector",
				"notarized by shard", shardID,
				"round", selfNotarizedHeaders[i].GetRound(),
				"nonce", selfNotarizedHeaders[i].GetNonce(),
				"hash", selfNotarizedHeadersHashes[i])

			selfNotarizedHeaderAdded = true
		}
	}

	return selfNotarizedHeaderAdded
}

func (sfd *shardForkDetector) computeFinalCheckpoint() {
	sfd.mutFinalityUpdate.Lock()
	sfd.computeFinalCheckpointLocked()
	sfd.mutFinalityUpdate.Unlock()
}

// ReconcileFinalCheckpoint serializes exact reconciliation with shard finality updates.
func (sfd *shardForkDetector) ReconcileFinalCheckpoint(nonce uint64) bool {
	sfd.mutFinalityUpdate.Lock()
	defer sfd.mutFinalityUpdate.Unlock()

	return sfd.baseForkDetector.ReconcileFinalCheckpoint(nonce)
}

// ReconcileFinalCheckpointBelow serializes suffix removal with shard finality updates.
func (sfd *shardForkDetector) ReconcileFinalCheckpointBelow(nonce uint64) bool {
	sfd.mutFinalityUpdate.Lock()
	reconciled, loweredFinal := sfd.reconcileFinalCheckpointRecordsBelow(nonce, nil)
	sfd.mutFinalityUpdate.Unlock()
	if reconciled {
		sfd.finishFinalCheckpointReconciliation(nonce, loweredFinal)
	}

	return reconciled
}

// ReconcileFinalCheckpointFromAuthority serializes suffix removal with shard finality updates.
func (sfd *shardForkDetector) ReconcileFinalCheckpointFromAuthority(nonce uint64, selectedHash []byte) bool {
	if len(selectedHash) == 0 {
		return false
	}

	sfd.mutFinalityUpdate.Lock()
	reconciled, loweredFinal := sfd.reconcileFinalCheckpointRecordsBelow(nonce, selectedHash)
	sfd.mutFinalityUpdate.Unlock()
	if reconciled {
		sfd.finishFinalCheckpointReconciliation(nonce, loweredFinal)
	}

	return reconciled
}

func (sfd *shardForkDetector) computeFinalCheckpointLocked() {
	finalCheckpoint := &checkpointInfo{}
	finalCheckpointWasSet := false
	finalCheckpointIsV3 := false

	sfd.mutHeaders.Lock()
	for nonce, headersInfo := range sfd.headers {
		if finalCheckpoint.nonce >= nonce {
			continue
		}

		indexBHProcessed, indexBHNotarized := sfd.getProcessedAndNotarizedIndexes(headersInfo)
		isProcessedBlockAlreadyNotarized := indexBHProcessed != -1 && indexBHNotarized != -1
		if !isProcessedBlockAlreadyNotarized {
			continue
		}
		if !headersInfo[indexBHProcessed].hasProof {
			continue
		}

		sameHash := bytes.Equal(headersInfo[indexBHNotarized].hash, headersInfo[indexBHProcessed].hash)
		if !sameHash {
			continue
		}

		finalCheckpoint = &checkpointInfo{
			nonce: nonce,
			round: headersInfo[indexBHNotarized].round,
			hash:  headersInfo[indexBHNotarized].hash,
		}

		finalCheckpointWasSet = true
		finalCheckpointIsV3 = sfd.isAsyncExecutionEnabled(headersInfo[indexBHProcessed])
	}

	if finalCheckpointWasSet {
		canAdvance := !finalCheckpointIsV3 || sfd.reconcileAuthorityAlongProcessedAncestryLocked(finalCheckpoint)
		if canAdvance {
			// a processed block matching its meta notarization is the settlement anchor
			sfd.advanceFinalAndSettledCheckpoint(finalCheckpoint)
		}
	}
	sfd.mutHeaders.Unlock()

	sfd.finalizeCleanProcessedDescendants()
	sfd.logFinalityLag()
}

func (sfd *shardForkDetector) reconcileAuthorityAlongProcessedAncestryLocked(candidate *checkpointInfo) bool {
	settled := sfd.settledCheckpoint()
	if candidate.nonce <= settled.nonce {
		return true
	}

	parentHash := settled.hash
	for nonce := settled.nonce + 1; ; nonce++ {
		processed := sfd.processedChildOnBranchLocked(nonce, parentHash)
		if processed == nil || (nonce == candidate.nonce && !bytes.Equal(processed.hash, candidate.hash)) {
			return false
		}
		if nonce == candidate.nonce {
			break
		}
		parentHash = processed.hash
	}

	parentHash = settled.hash
	removed := false
	for nonce := settled.nonce + 1; ; nonce++ {
		processed := sfd.processedChildOnBranchLocked(nonce, parentHash)
		selection := sfd.getNotarizedHeaderSelectionLocked(nonce)
		if selection.isV3 && len(selection.candidates) > 1 {
			retained := sfd.headers[nonce][:0]
			for _, info := range sfd.headers[nonce] {
				if info.state == process.BHNotarized && !bytes.Equal(info.hash, processed.hash) {
					removed = true
					continue
				}
				retained = append(retained, info)
			}
			sfd.headers[nonce] = retained
		}

		if nonce == candidate.nonce {
			break
		}
		parentHash = processed.hash
	}
	if removed {
		sfd.refreshAmbiguousNotarizationLocked()
	}

	return true
}

func (sfd *shardForkDetector) processedChildOnBranchLocked(nonce uint64, parentHash []byte) *headerInfo {
	var selected *headerInfo
	for _, info := range sfd.headers[nonce] {
		if info.state != process.BHProcessed || !bytes.Equal(info.prevHash, parentHash) {
			continue
		}
		if selected != nil && !bytes.Equal(selected.hash, info.hash) {
			return nil
		}
		selected = info
	}

	return selected
}

// finalizeCleanProcessedDescendants extends the final checkpoint over processed descendants that
// have a proof and no skipped round, so instant finality resumes after a settled contention
func (sfd *shardForkDetector) finalizeCleanProcessedDescendants() {
	finalCheckpoint := sfd.finalCheckpoint()
	if len(finalCheckpoint.hash) == 0 {
		return
	}

	for {
		sfd.mutHeaders.RLock()
		child := sfd.getCleanProcessedChild(finalCheckpoint)
		sfd.mutHeaders.RUnlock()
		if child == nil {
			break
		}
		if sfd.isAsyncExecutionEnabled(child) &&
			sfd.proofsPool.HasProofForDifferentHash(sfd.shardID, child.nonce, child.hash) {
			break
		}

		sfd.mutHeaders.RLock()
		confirmedChild := sfd.getCleanProcessedChild(finalCheckpoint)
		if confirmedChild == nil || !bytes.Equal(confirmedChild.hash, child.hash) {
			sfd.mutHeaders.RUnlock()
			break
		}

		finalCheckpoint = &checkpointInfo{
			nonce: confirmedChild.nonce,
			round: confirmedChild.round,
			hash:  confirmedChild.hash,
		}
		if sfd.isAsyncExecutionEnabled(confirmedChild) {
			sfd.advanceFinalCheckpoint(finalCheckpoint)
		} else {
			sfd.advanceFinalAndSettledCheckpoint(finalCheckpoint)
		}
		sfd.mutHeaders.RUnlock()
	}
}

func (sfd *shardForkDetector) getCleanProcessedChild(parent *checkpointInfo) *headerInfo {
	var processedChild *headerInfo
	for _, hdrInfo := range sfd.headers[parent.nonce+1] {
		isCleanProcessedChild := hdrInfo.state == process.BHProcessed &&
			hdrInfo.hasProof &&
			bytes.Equal(hdrInfo.prevHash, parent.hash) &&
			!common.IsContendedRound(hdrInfo.round, parent.round) &&
			sfd.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, hdrInfo.epoch)
		if isCleanProcessedChild {
			processedChild = hdrInfo
			break
		}
	}
	if processedChild == nil {
		return nil
	}
	if !sfd.isAsyncExecutionEnabled(processedChild) {
		return processedChild
	}

	if sfd.hasCompetingSiblingEvidenceLocked(
		processedChild.nonce,
		processedChild.hash,
		processedChild.prevHash,
	) {
		return nil
	}

	return processedChild
}

func (sfd *shardForkDetector) getProcessedAndNotarizedIndexes(headersInfo []*headerInfo) (int, int) {
	indexBHProcessed := -1
	indexBHNotarized := -1

	for index, hdrInfo := range headersInfo {
		switch hdrInfo.state {
		case process.BHProcessed:
			if indexBHProcessed != -1 && !bytes.Equal(headersInfo[indexBHProcessed].hash, hdrInfo.hash) &&
				(sfd.isAsyncExecutionEnabled(headersInfo[indexBHProcessed]) || sfd.isAsyncExecutionEnabled(hdrInfo)) {
				return -1, -1
			}
			indexBHProcessed = index
		case process.BHNotarized:
			if indexBHNotarized != -1 && !bytes.Equal(headersInfo[indexBHNotarized].hash, hdrInfo.hash) &&
				(sfd.isAsyncExecutionEnabled(headersInfo[indexBHNotarized]) || sfd.isAsyncExecutionEnabled(hdrInfo)) {
				return -1, -1
			}
			indexBHNotarized = index
		case process.BHReceived, process.BHReceivedTooLate:
			// legitimate coexisting entries, not relevant for the final checkpoint
		default:
			log.Error("invalid header state in fork detector", "state", hdrInfo.state, "nonce", hdrInfo.nonce, "round", hdrInfo.round, "hash", hdrInfo.hash)
		}
	}

	return indexBHProcessed, indexBHNotarized
}

// IsInterfaceNil returns true if there is no value under the interface
func (sfd *shardForkDetector) IsInterfaceNil() bool {
	return sfd == nil
}
