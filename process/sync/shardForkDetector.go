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

type invalidatedSelfNotarizedFromCrossHeadersRegistrar interface {
	RegisterInvalidatedSelfNotarizedFromCrossHeadersHandler(
		handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
	)
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
	bfd.setFinalCheckpoint(checkpoint)
	bfd.setSettledCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)
	bfd.fork.rollBackNonce = math.MaxUint64
	bfd.fork.probableHighestNonce = bfd.genesisNonce
	bfd.fork.highestNonceReceived = bfd.genesisNonce

	sfd := shardForkDetector{
		baseForkDetector: bfd,
	}

	sfd.blockTracker.RegisterSelfNotarizedFromCrossHeadersHandler(sfd.ReceivedSelfNotarizedFromCrossHeaders)
	if registrar, ok := blockTracker.(invalidatedSelfNotarizedFromCrossHeadersRegistrar); ok {
		registrar.RegisterInvalidatedSelfNotarizedFromCrossHeadersHandler(
			sfd.InvalidatedSelfNotarizedFromCrossHeaders,
		)
	}

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
		sfd.setFinalCheckpoint(newCheckpoint)
		// under Supernova the settled checkpoint advances only on meta notarization
		if !sfd.isSupernovaForHeader(header) {
			sfd.setSettledCheckpoint(newCheckpoint)
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

// InvalidatedSelfNotarizedFromCrossHeaders removes V3 shard authority derived from a dead meta branch.
func (sfd *shardForkDetector) InvalidatedSelfNotarizedFromCrossHeaders(
	shardID uint32,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) {
	if shardID != core.MetachainShardId {
		return
	}

	invalidated := make(map[uint64][][]byte)
	for index, header := range selfNotarizedHeaders {
		if index >= len(selfNotarizedHeadersHashes) || check.IfNil(header) || !header.IsHeaderV3() ||
			!common.IsCrossHeaderSettlementEnabledForHeader(sfd.enableEpochsHandler, sfd.enableRoundsHandler, header) {
			continue
		}

		invalidated[header.GetNonce()] = append(
			invalidated[header.GetNonce()],
			selfNotarizedHeadersHashes[index],
		)
	}
	if len(invalidated) == 0 {
		return
	}

	sfd.mutFinalityUpdate.Lock()

	removed := false
	sfd.mutHeaders.Lock()
	for nonce, hashes := range invalidated {
		headerInfos := sfd.headers[nonce]
		retained := make([]*headerInfo, 0, len(headerInfos))
		for _, headerInfo := range headerInfos {
			if headerInfo.state == process.BHNotarized && containsHash(hashes, headerInfo.hash) {
				removed = true
				continue
			}
			retained = append(retained, headerInfo)
		}
		if len(retained) == 0 {
			delete(sfd.headers, nonce)
			continue
		}
		sfd.headers[nonce] = retained
	}
	if removed {
		sfd.refreshAmbiguousNotarizationLocked()
	}
	sfd.mutHeaders.Unlock()
	if !removed {
		sfd.mutFinalityUpdate.Unlock()
		return
	}

	sfd.lowerInvalidatedCheckpoints(invalidated)
	sfd.computeFinalCheckpointLocked()
	sfd.mutFinalityUpdate.Unlock()
	sfd.recomputeProbableHighestNonce()
}

func containsHash(hashes [][]byte, hash []byte) bool {
	for _, candidate := range hashes {
		if bytes.Equal(candidate, hash) {
			return true
		}
	}

	return false
}

func (sfd *shardForkDetector) lowerInvalidatedCheckpoints(invalidated map[uint64][][]byte) {
	sfd.mutFork.Lock()
	defer sfd.mutFork.Unlock()

	retainedHistory := sfd.fork.settledCheckpointHistory[:0]
	for _, checkpoint := range sfd.fork.settledCheckpointHistory {
		if !checkpointWasInvalidated(checkpoint, invalidated) {
			retainedHistory = append(retainedHistory, checkpoint)
		}
	}
	sfd.fork.settledCheckpointHistory = retainedHistory

	if checkpointWasInvalidated(sfd.fork.settledCheckpoint, invalidated) {
		sfd.fork.settledCheckpoint = sfd.lastSettledCheckpointLocked()
	}
	if checkpointWasInvalidated(sfd.fork.finalCheckpoint, invalidated) {
		sfd.fork.finalCheckpoint = sfd.highestProcessedCheckpointBelowLocked(sfd.fork.finalCheckpoint.nonce)
	}
}

func checkpointWasInvalidated(checkpoint *checkpointInfo, invalidated map[uint64][][]byte) bool {
	if checkpoint == nil {
		return false
	}

	return containsHash(invalidated[checkpoint.nonce], checkpoint.hash)
}

func (sfd *shardForkDetector) lastSettledCheckpointLocked() *checkpointInfo {
	lastIndex := len(sfd.fork.settledCheckpointHistory) - 1
	if lastIndex < 0 {
		return &checkpointInfo{
			nonce: sfd.genesisNonce,
			round: sfd.genesisRound,
		}
	}

	checkpoint := sfd.fork.settledCheckpointHistory[lastIndex]
	sfd.fork.settledCheckpointHistory = sfd.fork.settledCheckpointHistory[:lastIndex]

	return checkpoint
}

func (sfd *shardForkDetector) highestProcessedCheckpointBelowLocked(nonce uint64) *checkpointInfo {
	checkpoint := sfd.fork.settledCheckpoint
	if checkpoint == nil {
		checkpoint = &checkpointInfo{
			nonce: sfd.genesisNonce,
			round: sfd.genesisRound,
		}
	}
	for _, candidate := range sfd.fork.checkpoint {
		if candidate.nonce < nonce && candidate.nonce >= checkpoint.nonce {
			checkpoint = candidate
		}
	}

	return checkpoint
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

func (sfd *shardForkDetector) computeFinalCheckpointLocked() {
	finalCheckpoint := &checkpointInfo{}
	finalCheckpointWasSet := false

	sfd.mutHeaders.RLock()
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
	}
	sfd.mutHeaders.RUnlock()

	if finalCheckpointWasSet {
		sfd.advanceFinalCheckpoint(finalCheckpoint)
		// a processed block matching its meta notarization is the settlement anchor
		sfd.advanceSettledCheckpoint(finalCheckpoint)
	}

	sfd.finalizeCleanProcessedDescendants()
	sfd.logFinalityLag()
}

// finalizeCleanProcessedDescendants extends the final checkpoint over processed descendants that
// have a proof and no skipped round, so instant finality resumes after a settled contention
func (sfd *shardForkDetector) finalizeCleanProcessedDescendants() {
	finalCheckpoint := sfd.finalCheckpoint()
	if len(finalCheckpoint.hash) == 0 {
		return
	}

	advanced := false
	sfd.mutHeaders.RLock()
	for {
		child := sfd.getCleanProcessedChild(finalCheckpoint)
		if child == nil {
			break
		}

		finalCheckpoint = &checkpointInfo{nonce: child.nonce, round: child.round, hash: child.hash}
		advanced = true
	}
	sfd.mutHeaders.RUnlock()

	if advanced {
		sfd.advanceFinalCheckpoint(finalCheckpoint)
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
