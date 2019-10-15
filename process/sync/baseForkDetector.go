package sync

import (
	"bytes"
	"math"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

const MinForkRound = uint64(0)

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
}

// baseForkDetector defines a struct with necessary data needed for fork detection
type baseForkDetector struct {
	rounder consensus.Rounder

	headers    map[uint64][]*headerInfo
	mutHeaders sync.RWMutex
	fork       forkInfo
	mutFork    sync.RWMutex
}

func (bfd *baseForkDetector) removePastOrInvalidRecords() {
	bfd.removePastHeaders()
	bfd.removeInvalidReceivedHeaders()
	bfd.removePastCheckpoints()
}

func (bfd *baseForkDetector) checkBlockBasicValidity(header data.HeaderHandler, state process.BlockHeaderState) error {
	roundDif := int64(header.GetRound()) - int64(bfd.finalCheckpoint().round)
	nonceDif := int64(header.GetNonce()) - int64(bfd.finalCheckpoint().nonce)
	//TODO: Analyze if the acceptance of some headers which came for the next round could generate some attack vectors
	nextRound := bfd.rounder.Index() + 1

	if roundDif <= 0 {
		return ErrLowerRoundInBlock
	}
	if nonceDif <= 0 {
		return ErrLowerNonceInBlock
	}
	if int64(header.GetRound()) > nextRound {
		return ErrHigherRoundInBlock
	}
	if roundDif < nonceDif {
		return ErrHigherNonceInBlock
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
		if bytes.Equal(hdrInfoStored.hash, hash) {
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
		if bytes.Equal(hdrInfoStored.hash, hdrInfo.hash) {
			if hdrInfoStored.state != process.BHProcessed {
				// If the old appended header has the same hash with the new one received, than the state of the old
				// record will be replaced if the new one is more important. Below is the hierarchy, from low to high,
				// of the record state importance: (BHProposed, BHReceived, BHNotarized, BHProcessed)
				if hdrInfo.state == process.BHNotarized || hdrInfo.state == process.BHProcessed {
					hdrInfoStored.state = hdrInfo.state
				}
			}
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

// ResetProbableHighestNonceIfNeeded resets the probableHighestNonce to checkpoint if after maxRoundsToWait nothing
// is received so the node will act as synchronized
func (bfd *baseForkDetector) ResetProbableHighestNonceIfNeeded() {
	//TODO: This mechanism should be improved to avoid the situation when a malicious group of 2/3 + 1 from a
	// consensus group size, could keep all the shard in sync mode, by creating fake blocks higher than current
	// committed block + 1, which could not be verified by hash -> prev hash and only by rand seed -> prev random seed
	roundsWithoutReceivedBlock := bfd.rounder.Index() - int64(bfd.lastBlockRound())
	if roundsWithoutReceivedBlock > maxRoundsToWait {
		probableHighestNonce := bfd.ProbableHighestNonce()
		checkpointNonce := bfd.lastCheckpoint().nonce
		if probableHighestNonce > checkpointNonce {
			bfd.setProbableHighestNonce(checkpointNonce)
		}
	}
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

// IsInterfaceNil returns true if there is no value under the interface
func (bfd *baseForkDetector) IsInterfaceNil() bool {
	if bfd == nil {
		return true
	}
	return false
}

// CheckFork method checks if the node could be on the fork
func (bfd *baseForkDetector) CheckFork() (bool, uint64, []byte) {
	var (
		lowestForkNonce        uint64
		hashOfLowestForkNonce  []byte
		lowestRoundInForkNonce uint64
		forkHeaderHash         []byte
		selfHdrInfo            *headerInfo
	)

	lowestForkNonce = math.MaxUint64
	hashOfLowestForkNonce = nil
	forkDetected := false

	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		if len(hdrInfos) == 1 {
			continue
		}

		selfHdrInfo = nil
		lowestRoundInForkNonce = math.MaxUint64
		forkHeaderHash = nil

		for i := 0; i < len(hdrInfos); i++ {
			if hdrInfos[i].state == process.BHProcessed {
				selfHdrInfo = hdrInfos[i]
				continue
			}

			forkHeaderHash, lowestRoundInForkNonce = bfd.computeForkInfo(
				hdrInfos[i],
				forkHeaderHash,
				lowestRoundInForkNonce)
		}

		if selfHdrInfo == nil {
			// if current nonce has not been processed yet, then skip and check the next one.
			continue
		}

		if bfd.shouldSignalFork(selfHdrInfo, forkHeaderHash, lowestRoundInForkNonce) {
			forkDetected = true
			if nonce < lowestForkNonce {
				lowestForkNonce = nonce
				hashOfLowestForkNonce = forkHeaderHash
			}
			continue
		}

		// keep it clean so next time this position will be processed faster
		delete(bfd.headers, nonce)
		bfd.headers[nonce] = []*headerInfo{selfHdrInfo}
	}
	bfd.mutHeaders.Unlock()

	return forkDetected, lowestForkNonce, hashOfLowestForkNonce
}

func (bfd *baseForkDetector) computeForkInfo(
	headerInfo *headerInfo,
	lastForkHash []byte,
	lastForkRound uint64,
) ([]byte, uint64) {

	currentForkRound := headerInfo.round
	if headerInfo.state == process.BHNotarized {
		currentForkRound = MinForkRound
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

	higherHashForSameRound := headerInfo.round == lastForkRound &&
		strings.Compare(string(headerInfo.hash), string(lastForkHash)) > 0
	shouldSignalFork := headerInfo.round > lastForkRound || higherHashForSameRound

	return shouldSignalFork
}

func (bfd *baseForkDetector) shouldAddBlockInForkDetector(
	header data.HeaderHandler,
	state process.BlockHeaderState,
	finality int64,
) error {

	noncesDifference := int64(bfd.ProbableHighestNonce()) - int64(header.GetNonce())
	isSyncing := state == process.BHReceived && noncesDifference > process.MaxNoncesDifference
	if state == process.BHProcessed || isSyncing {
		return nil
	}

	roundTooOld := int64(header.GetRound()) < bfd.rounder.Index()-finality
	if roundTooOld {
		return ErrLowerRoundInBlock
	}

	return nil
}
