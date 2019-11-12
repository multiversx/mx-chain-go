package sync

import (
	"bytes"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type headerInfo struct {
	nonce uint64
	round uint64
	hash  []byte
	state process.BlockHeaderState
}

type checkpointInfo struct {
	nonce uint64
	round uint64
}

type forkInfo struct {
	checkpoint             []*checkpointInfo
	finalCheckpoint        *checkpointInfo
	probableHighestNonce   uint64
	lastBlockRound         uint64
	lastProposedBlockNonce uint64
	shouldForceFork        bool
	isNotarizedShardStuck  bool
}

// baseForkDetector defines a struct with necessary data needed for fork detection
type baseForkDetector struct {
	rounder consensus.Rounder

	headers    map[uint64][]*headerInfo
	mutHeaders sync.RWMutex
	fork       forkInfo
	mutFork    sync.RWMutex

	blackListHandler process.BlackListHandler
}

func (bfd *baseForkDetector) removePastOrInvalidRecords() {
	bfd.removePastHeaders()
	bfd.removeInvalidReceivedHeaders()
	bfd.removePastCheckpoints()
}

func (bfd *baseForkDetector) checkBlockBasicValidity(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
) error {

	roundDif := int64(header.GetRound()) - int64(bfd.finalCheckpoint().round)
	nonceDif := int64(header.GetNonce()) - int64(bfd.finalCheckpoint().nonce)
	//TODO: Analyze if the acceptance of some headers which came for the next round could generate some attack vectors
	nextRound := bfd.rounder.Index() + 1

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
	if bfd.blackListHandler.Has(string(header.GetPrevHash())) {
		//TODO: Should be done some tests to reconsider adding here to the black list also this received header,
		// which is bound to a previous black listed header.
		bfd.blackListHandler.Add(string(headerHash))
		return process.ErrHeaderIsBlackListed
	}
	if state == process.BHProposed {
		if !isRandomSeedValid(header) {
			return ErrRandomSeedNotValid
		}
	}
	if state == process.BHReceived || state == process.BHProcessed {
		if !isSigned(header) {
			return ErrBlockIsNotSigned
		}
	}

	return nil
}

func (bfd *baseForkDetector) removePastHeaders() {
	finalCheckpointNonce := bfd.finalCheckpoint().nonce

	bfd.mutHeaders.Lock()
	for nonce := range bfd.headers {
		if nonce < finalCheckpointNonce {
			delete(bfd.headers, nonce)
		}
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *baseForkDetector) removeInvalidReceivedHeaders() {
	finalCheckpointRound := bfd.finalCheckpoint().round
	finalCheckpointNonce := bfd.finalCheckpoint().nonce

	var validHdrInfos []*headerInfo

	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		validHdrInfos = nil
		for i := 0; i < len(hdrInfos); i++ {
			roundDif := int64(hdrInfos[i].round) - int64(finalCheckpointRound)
			nonceDif := int64(hdrInfos[i].nonce) - int64(finalCheckpointNonce)
			isReceivedHeaderInvalid := hdrInfos[i].state == process.BHReceived && roundDif < nonceDif
			if isReceivedHeaderInvalid {
				continue
			}

			validHdrInfos = append(validHdrInfos, hdrInfos[i])
		}
		if validHdrInfos == nil {
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
	var preservedCheckpoint []*checkpointInfo

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
	probableHighestNonce := bfd.finalCheckpoint().nonce
	lastProposedBlockNonce := bfd.lastProposedBlockNonce()
	if lastProposedBlockNonce > 0 {
		probableHighestNonce = core.MaxUint64(bfd.finalCheckpoint().nonce, lastProposedBlockNonce-1)
	}

	bfd.mutHeaders.RLock()
	for nonce := range bfd.headers {
		if nonce <= probableHighestNonce {
			continue
		}
		probableHighestNonce = nonce
	}
	bfd.mutHeaders.RUnlock()

	return probableHighestNonce
}

// RemoveHeaders removes all the stored headers with a given nonce
func (bfd *baseForkDetector) RemoveHeaders(nonce uint64, hash []byte) {
	bfd.removeCheckpointWithNonce(nonce)

	var preservedHdrInfos []*headerInfo

	bfd.mutHeaders.Lock()
	hdrInfos := bfd.headers[nonce]
	for _, hdrInfoStored := range hdrInfos {
		if hdrInfoStored.state != process.BHNotarized {
			continue
		}

		preservedHdrInfos = append(preservedHdrInfos, hdrInfoStored)
	}

	if preservedHdrInfos == nil {
		delete(bfd.headers, nonce)
	} else {
		bfd.headers[nonce] = preservedHdrInfos
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *baseForkDetector) removeCheckpointWithNonce(nonce uint64) {
	bfd.mutFork.Lock()
	var preservedCheckpoint []*checkpointInfo

	for i := 0; i < len(bfd.fork.checkpoint); i++ {
		if bfd.fork.checkpoint[i].nonce == nonce {
			continue
		}

		preservedCheckpoint = append(preservedCheckpoint, bfd.fork.checkpoint[i])
	}

	bfd.fork.checkpoint = preservedCheckpoint
	bfd.mutFork.Unlock()
}

// append adds a new header in the slice found in nonce position
// it not adds the header if its hash is already stored in the slice
func (bfd *baseForkDetector) append(hdrInfo *headerInfo) {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	// Proposed blocks received do not count for fork choice, as they are not valid until the consensus
	// is achieved. They should be received afterwards through sync mechanism.
	if hdrInfo.state == process.BHProposed {
		bfd.setLastProposedBlockNonce(hdrInfo.nonce)
		return
	}

	hdrInfos := bfd.headers[hdrInfo.nonce]
	isHdrInfosNilOrEmpty := hdrInfos == nil || len(hdrInfos) == 0
	if isHdrInfosNilOrEmpty {
		bfd.headers[hdrInfo.nonce] = []*headerInfo{hdrInfo}
		return
	}

	for _, hdrInfoStored := range hdrInfos {
		if bytes.Equal(hdrInfoStored.hash, hdrInfo.hash) && hdrInfoStored.state == hdrInfo.state {
			return
		}
	}

	bfd.headers[hdrInfo.nonce] = append(bfd.headers[hdrInfo.nonce], hdrInfo)
}

// GetHighestFinalBlockNonce gets the highest nonce of the block which is final and it can not be reverted anymore
func (bfd *baseForkDetector) GetHighestFinalBlockNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

// ProbableHighestNonce gets the probable highest nonce
func (bfd *baseForkDetector) ProbableHighestNonce() uint64 {
	return bfd.probableHighestNonce()
}

// ResetProbableHighestNonce resets the probableHighestNonce to checkpoint
func (bfd *baseForkDetector) ResetProbableHighestNonce() {
	bfd.setProbableHighestNonce(bfd.lastCheckpoint().nonce)
}

// ResetFork resets the forced fork
func (bfd *baseForkDetector) ResetFork() {
	bfd.cleanupReceivedHeadersHigherThanNonce(bfd.lastCheckpoint().nonce)
	bfd.setShouldForceFork(false)
}

func (bfd *baseForkDetector) addCheckpoint(checkpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	bfd.fork.checkpoint = append(bfd.fork.checkpoint, checkpoint)
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) lastCheckpoint() *checkpointInfo {
	bfd.mutFork.RLock()
	lastIndex := len(bfd.fork.checkpoint) - 1
	if lastIndex < 0 {
		bfd.mutFork.RUnlock()
		return &checkpointInfo{}
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

func (bfd *baseForkDetector) finalCheckpoint() *checkpointInfo {
	bfd.mutFork.RLock()
	finalCheckpoint := bfd.fork.finalCheckpoint
	bfd.mutFork.RUnlock()

	return finalCheckpoint
}

func (bfd *baseForkDetector) setProbableHighestNonce(nonce uint64) {
	bfd.mutFork.Lock()
	bfd.fork.probableHighestNonce = nonce
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) probableHighestNonce() uint64 {
	bfd.mutFork.RLock()
	probableHighestNonce := bfd.fork.probableHighestNonce
	bfd.mutFork.RUnlock()

	return probableHighestNonce
}

func (bfd *baseForkDetector) setLastBlockRound(round uint64) {
	bfd.mutFork.Lock()
	bfd.fork.lastBlockRound = round
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) lastBlockRound() uint64 {
	bfd.mutFork.RLock()
	lastBlockRound := bfd.fork.lastBlockRound
	bfd.mutFork.RUnlock()

	return lastBlockRound
}

func (bfd *baseForkDetector) setLastProposedBlockNonce(nonce uint64) {
	bfd.mutFork.Lock()
	bfd.fork.lastProposedBlockNonce = nonce
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) lastProposedBlockNonce() uint64 {
	bfd.mutFork.RLock()
	lastProposedBlockNonce := bfd.fork.lastProposedBlockNonce
	bfd.mutFork.RUnlock()

	return lastProposedBlockNonce
}

func (bfd *baseForkDetector) setShouldForceFork(shouldForceFork bool) {
	bfd.mutFork.Lock()
	bfd.fork.shouldForceFork = shouldForceFork
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) shouldForceFork() bool {
	bfd.mutFork.RLock()
	shouldForceFork := bfd.fork.shouldForceFork
	bfd.mutFork.RUnlock()

	return shouldForceFork
}

func (bfd *baseForkDetector) setIsNotarizedShardStuck(isNotarizedShardStuck bool) {
	bfd.mutFork.Lock()
	bfd.fork.isNotarizedShardStuck = isNotarizedShardStuck
	bfd.mutFork.Unlock()
}

func (bfd *baseForkDetector) isNotarizedShardStuck() bool {
	bfd.mutFork.RLock()
	isNotarizedShardStuck := bfd.fork.isNotarizedShardStuck
	bfd.mutFork.RUnlock()

	return isNotarizedShardStuck
}

// IsInterfaceNil returns true if there is no value under the interface
func (bfd *baseForkDetector) IsInterfaceNil() bool {
	if bfd == nil {
		return true
	}
	return false
}

// CheckFork method checks if the node could be on the fork
func (bfd *baseForkDetector) CheckFork() *process.ForkInfo {
	var (
		forkHeaderRound uint64
		forkHeaderHash  []byte
		selfHdrInfo     *headerInfo
	)

	forkInfo := process.NewForkInfo()

	if bfd.shouldForceFork() {
		forkInfo.IsDetected = true
		return forkInfo
	}

	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		if len(hdrInfos) == 1 {
			continue
		}

		selfHdrInfo = nil
		forkHeaderRound = math.MaxUint64
		forkHeaderHash = nil

		for i := 0; i < len(hdrInfos); i++ {
			if hdrInfos[i].state == process.BHProcessed {
				selfHdrInfo = hdrInfos[i]
				continue
			}

			forkHeaderHash, forkHeaderRound = bfd.computeForkInfo(
				hdrInfos[i],
				forkHeaderHash,
				forkHeaderRound)
		}

		if selfHdrInfo == nil {
			// if current nonce has not been processed yet, then skip and check the next one.
			continue
		}

		if bfd.shouldSignalFork(selfHdrInfo, forkHeaderHash, forkHeaderRound) {
			forkInfo.IsDetected = true
			if nonce < forkInfo.Nonce {
				forkInfo.Nonce = nonce
				forkInfo.Round = forkHeaderRound
				forkInfo.Hash = forkHeaderHash
			}
		}
	}
	bfd.mutHeaders.Unlock()

	return forkInfo
}

func (bfd *baseForkDetector) computeForkInfo(
	headerInfo *headerInfo,
	lastForkHash []byte,
	lastForkRound uint64,
) ([]byte, uint64) {

	currentForkRound := headerInfo.round
	if headerInfo.state == process.BHNotarized {
		currentForkRound = process.MinForkRound
	}

	if currentForkRound < lastForkRound {
		return headerInfo.hash, currentForkRound
	}

	lowerHashForSameRound := currentForkRound == lastForkRound &&
		bytes.Compare(headerInfo.hash, lastForkHash) < 0
	if lowerHashForSameRound {
		return headerInfo.hash, currentForkRound
	}

	return lastForkHash, lastForkRound
}

func (bfd *baseForkDetector) shouldSignalFork(
	headerInfo *headerInfo,
	lastForkHash []byte,
	lastForkRound uint64,
) bool {
	sameHash := bytes.Equal(headerInfo.hash, lastForkHash)
	higherHashForSameRound := headerInfo.round == lastForkRound &&
		bytes.Compare(headerInfo.hash, lastForkHash) > 0
	shouldSignalFork := !sameHash && (headerInfo.round > lastForkRound || higherHashForSameRound)

	return shouldSignalFork
}

func (bfd *baseForkDetector) shouldAddBlockInForkDetector(
	header data.HeaderHandler,
	state process.BlockHeaderState,
	finality int64,
) error {

	if state == process.BHProcessed {
		return nil
	}

	// This condition would avoid a stuck situation, when shards would set as final, block with nonce n received from
	// meta-chain, because they also received n+1. In the same time meta-chain would be reverted to an older block with
	// nonce n received it with latency but before n+1. Actually this condition would reject these older blocks.
	roundTooOld := int64(header.GetRound()) < bfd.rounder.Index()-finality
	if roundTooOld {
		return ErrLowerRoundInBlock
	}

	return nil
}

func (bfd *baseForkDetector) activateForcedForkIfNeeded(
	header data.HeaderHandler,
	state process.BlockHeaderState,
) {

	if state != process.BHProposed || bfd.isSyncing() {
		return
	}

	lastCheckpointRound := bfd.lastCheckpoint().round
	lastCheckpointNonce := bfd.lastCheckpoint().nonce
	finalCheckpointNonce := bfd.finalCheckpoint().nonce

	roundsDifference := int64(header.GetRound()) - int64(lastCheckpointRound)
	noncesDifference := int64(header.GetNonce()) - int64(lastCheckpointNonce)
	noncesWithoutCrossNotarizedDifference := int64(header.GetNonce()) - int64(finalCheckpointNonce)
	isInProperRound := process.IsInProperRound(bfd.rounder.Index())

	isConsensusStuck := roundsDifference > process.MaxRoundsWithoutCommittedBlock &&
		noncesDifference <= 1 &&
		isInProperRound

	isCrossNotarizedStuck := noncesWithoutCrossNotarizedDifference > process.MaxNoncesWithoutCrossNotarized &&
		noncesDifference <= 1 &&
		isInProperRound &&
		!bfd.isNotarizedShardStuck()

	if isConsensusStuck || isCrossNotarizedStuck {
		bfd.setShouldForceFork(true)
	}
}

func (bfd *baseForkDetector) isSyncing() bool {
	noncesDifference := int64(bfd.ProbableHighestNonce()) - int64(bfd.lastCheckpoint().nonce)
	isSyncing := noncesDifference > process.NonceDifferenceWhenSynced
	return isSyncing
}

// GetNotarizedHeaderHash returns the hash of the header with a given nonce, if it has been received with state notarized
func (bfd *baseForkDetector) GetNotarizedHeaderHash(nonce uint64) []byte {
	bfd.mutHeaders.RLock()
	defer bfd.mutHeaders.RUnlock()

	hdrInfos := bfd.headers[nonce]
	for _, hdrInfo := range hdrInfos {
		if hdrInfo.state == process.BHNotarized {
			return hdrInfo.hash
		}
	}

	return nil
}

func (bfd *baseForkDetector) cleanupReceivedHeadersHigherThanNonce(nonce uint64) {
	bfd.mutHeaders.Lock()
	for hdrNonce, hdrInfos := range bfd.headers {
		if hdrNonce <= nonce {
			continue
		}

		preservedHdrInfos := make([]*headerInfo, 0)

		for _, hdrInfo := range hdrInfos {
			if hdrInfo.state != process.BHNotarized {
				continue
			}

			preservedHdrInfos = append(preservedHdrInfos, hdrInfo)
		}

		if len(preservedHdrInfos) == 0 {
			delete(bfd.headers, hdrNonce)
			continue
		}

		bfd.headers[hdrNonce] = preservedHdrInfos
	}
	bfd.mutHeaders.Unlock()
}
