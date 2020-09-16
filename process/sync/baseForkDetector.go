package sync

import (
	"bytes"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type headerInfo struct {
	epoch uint32
	nonce uint64
	round uint64
	hash  []byte
	state process.BlockHeaderState
}

type checkpointInfo struct {
	nonce uint64
	round uint64
	hash  []byte
}

type forkInfo struct {
	checkpoint              []*checkpointInfo
	finalCheckpoint         *checkpointInfo
	probableHighestNonce    uint64
	highestNonceReceived    uint64
	rollBackNonce           uint64
	lastRoundWithForcedFork int64
}

// baseForkDetector defines a struct with necessary data needed for fork detection
type baseForkDetector struct {
	rounder consensus.Rounder

	headers    map[uint64][]*headerInfo
	mutHeaders sync.RWMutex
	fork       forkInfo
	mutFork    sync.RWMutex

	blackListHandler   process.TimeCacher
	genesisTime        int64
	blockTracker       process.BlockTracker
	forkDetector       forkDetector
	genesisNonce       uint64
	genesisRound       uint64
	maxForkHeaderEpoch uint32
	genesisEpoch       uint32
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
	//TODO: Analyze if the acceptance of some headers which came for the next round could generate some attack vectors
	nextRound := bfd.rounder.Index() + 1
	genesisTimeFromHeader := bfd.computeGenesisTimeFromHeader(header)

	bfd.blackListHandler.Sweep()
	if bfd.blackListHandler.Has(string(header.GetPrevHash())) {
		process.AddHeaderToBlackList(bfd.blackListHandler, headerHash)
		return process.ErrHeaderIsBlackListed
	}
	//TODO: This check could be removed when this protection mechanism would be implemented on interceptors side
	if genesisTimeFromHeader != bfd.genesisTime {
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
	probableHighestNonce := bfd.finalCheckpoint().nonce

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

// RemoveHeader removes the stored header with the given nonce and hash
func (bfd *baseForkDetector) RemoveHeader(nonce uint64, hash []byte) {
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

	probableHighestNonce := bfd.computeProbableHighestNonce()
	bfd.setProbableHighestNonce(probableHighestNonce)

	log.Debug("forkDetector.RemoveHeader",
		"nonce", nonce,
		"hash", hash,
		"probable highest nonce", probableHighestNonce,
		"final check point nonce", bfd.finalCheckpoint().nonce)
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
		"last check point nonce", bfd.lastCheckpoint().nonce)
}

// append adds a new header in the slice found in nonce position
// it not adds the header if its hash is already stored in the slice
func (bfd *baseForkDetector) append(hdrInfo *headerInfo) bool {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	hdrInfos := bfd.headers[hdrInfo.nonce]
	isHdrInfosNilOrEmpty := len(hdrInfos) == 0 // no need for nil check, len() for nil returns 0
	if isHdrInfosNilOrEmpty {
		bfd.headers[hdrInfo.nonce] = []*headerInfo{hdrInfo}
		return true
	}

	for _, hdrInfoStored := range hdrInfos {
		if bytes.Equal(hdrInfoStored.hash, hdrInfo.hash) && hdrInfoStored.state == hdrInfo.state {
			return false
		}
	}

	bfd.headers[hdrInfo.nonce] = append(bfd.headers[hdrInfo.nonce], hdrInfo)
	return true
}

// GetHighestFinalBlockNonce gets the highest nonce of the block which is final and it can not be reverted anymore
func (bfd *baseForkDetector) GetHighestFinalBlockNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

// GetHighestFinalBlockHash gets the hash of the block which is final and it can not be reverted anymore
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
	bfd.setLastRoundWithForcedFork(bfd.rounder.Index())

	log.Debug("forkDetector.ResetFork",
		"last round with forced fork", bfd.lastRoundWithForcedFork())
}

// ResetProbableHighestNonce resets the probable highest nonce to the last checkpoint nonce / highest notarized nonce
func (bfd *baseForkDetector) ResetProbableHighestNonce() {
	bfd.cleanupReceivedHeadersHigherThanNonce(bfd.lastCheckpoint().nonce)
	probableHighestNonce := bfd.computeProbableHighestNonce()
	bfd.setProbableHighestNonce(probableHighestNonce)

	log.Debug("forkDetector.ResetProbableHighestNonce",
		"probable highest nonce", bfd.probableHighestNonce())
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

// RestoreToGenesis sets class variables to theirs initial values
func (bfd *baseForkDetector) RestoreToGenesis() {
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
	bfd.fork.probableHighestNonce = nonce
	bfd.mutFork.Unlock()
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

		selfHdrInfo = nil
		forkHeaderRound = math.MaxUint64
		forkHeaderHash = nil
		forkHeaderEpoch = 0
		bfd.maxForkHeaderEpoch = getMaxEpochFromHdrsInfo(hdrsInfo)

		for i := 0; i < len(hdrsInfo); i++ {
			if hdrsInfo[i].state == process.BHProcessed {
				selfHdrInfo = hdrsInfo[i]
				continue
			}

			forkHeaderHash, forkHeaderRound, forkHeaderEpoch = bfd.computeForkInfo(
				hdrsInfo[i],
				forkHeaderHash,
				forkHeaderRound,
				forkHeaderEpoch,
			)
		}

		if selfHdrInfo == nil {
			// if current nonce has not been processed yet, then skip and check the next one.
			continue
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

func getMaxEpochFromHdrsInfo(hdrInfos []*headerInfo) uint32 {
	maxEpoch := uint32(0)
	for _, hdrInfo := range hdrInfos {
		if hdrInfo.epoch > maxEpoch {
			maxEpoch = hdrInfo.epoch
		}
	}
	return maxEpoch
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

	if currentForkRound < lastForkRound {
		return hdrInfo.hash, currentForkRound, hdrInfo.epoch
	}

	lowerHashForSameRound := currentForkRound == lastForkRound &&
		bytes.Compare(hdrInfo.hash, lastForkHash) < 0
	if lowerHashForSameRound {
		return hdrInfo.hash, currentForkRound, hdrInfo.epoch
	}

	return lastForkHash, lastForkRound, lastForkEpoch
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
	isHeaderReceivedTooLate := int64(header.GetRound()) < bfd.rounder.Index()-finality

	return isHeaderReceivedTooLate
}

func (bfd *baseForkDetector) isConsensusStuck() bool {
	if bfd.lastRoundWithForcedFork() == bfd.rounder.Index() {
		return false
	}

	if bfd.isSyncing() {
		return false
	}

	roundsDifference := bfd.rounder.Index() - int64(bfd.lastCheckpoint().round)
	if roundsDifference <= process.MaxRoundsWithoutCommittedBlock {
		return false
	}

	if !process.IsInProperRound(bfd.rounder.Index()) {
		return false
	}

	return true
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
	for hdrsNonce, hdrsInfo := range bfd.headers {
		if hdrsNonce <= nonce {
			continue
		}

		preservedHdrsInfo := make([]*headerInfo, 0)

		for _, hdrInfo := range hdrsInfo {
			if hdrInfo.state != process.BHNotarized {
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

func (bfd *baseForkDetector) computeGenesisTimeFromHeader(headerHandler data.HeaderHandler) int64 {
	genesisTime := int64(headerHandler.GetTimeStamp() - (headerHandler.GetRound()-bfd.genesisRound)*uint64(bfd.rounder.TimeDuration().Seconds()))
	return genesisTime
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

func (bfd *baseForkDetector) processReceivedBlock(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
	doJobOnBHProcessed func(data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte),
) {
	bfd.setHighestNonceReceived(header.GetNonce())

	if state == process.BHProposed {
		return
	}

	isHeaderReceivedTooLate := bfd.isHeaderReceivedTooLate(header, state, process.BlockFinality)
	if isHeaderReceivedTooLate {
		state = process.BHReceivedTooLate
	}

	appended := bfd.append(&headerInfo{
		epoch: header.GetEpoch(),
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  headerHash,
		state: state,
	})
	if !appended {
		return
	}

	if state == process.BHProcessed {
		doJobOnBHProcessed(header, headerHash, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}

	probableHighestNonce := bfd.computeProbableHighestNonce()
	bfd.setProbableHighestNonce(probableHighestNonce)

	log.Debug("forkDetector.AddHeader",
		"round", header.GetRound(),
		"nonce", header.GetNonce(),
		"hash", headerHash,
		"state", state,
		"probable highest nonce", bfd.probableHighestNonce(),
		"last check point nonce", bfd.lastCheckpoint().nonce,
		"final check point nonce", bfd.finalCheckpoint().nonce)
}

// SetFinalToLastCheckpoint sets the final checkpoint to the last checkpoint added in list
func (bfd *baseForkDetector) SetFinalToLastCheckpoint() {
	bfd.setFinalCheckpoint(bfd.lastCheckpoint())
}
