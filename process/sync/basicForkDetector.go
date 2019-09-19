package sync

import (
	"bytes"
	"math"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
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
	checkpoint           []*checkpointInfo
	finalCheckpoint      *checkpointInfo
	probableHighestNonce uint64
	lastBlockRound       uint64
}

// basicForkDetector defines a struct with necessary data needed for fork detection
type basicForkDetector struct {
	rounder consensus.Rounder

	headers    map[uint64][]*headerInfo
	mutHeaders sync.RWMutex
	fork       forkInfo
	mutFork    sync.RWMutex
}

// NewBasicForkDetector method creates a new BasicForkDetector object
func NewBasicForkDetector(rounder consensus.Rounder,
) (*basicForkDetector, error) {
	if rounder == nil || rounder.IsInterfaceNil() {
		return nil, process.ErrNilRounder
	}

	bfd := &basicForkDetector{
		rounder: rounder,
	}

	bfd.headers = make(map[uint64][]*headerInfo)
	checkpoint := &checkpointInfo{}
	bfd.setFinalCheckpoint(checkpoint)
	bfd.addCheckpoint(checkpoint)

	return bfd, nil
}

// AddHeader method adds a new header to headers map
func (bfd *basicForkDetector) AddHeader(
	header data.HeaderHandler,
	headerHash []byte,
	state process.BlockHeaderState,
	finalHeaders []data.HeaderHandler,
	finalHeadersHashes [][]byte,
) error {

	if header == nil || header.IsInterfaceNil() {
		return ErrNilHeader
	}
	if headerHash == nil {
		return ErrNilHash
	}

	err := bfd.checkBlockValidity(header, state)
	if err != nil {
		return err
	}

	if state == process.BHProcessed {
		// add final header(s), add check point and remove all the past checkpoints/headers
		bfd.addFinalHeaders(finalHeaders, finalHeadersHashes)
		bfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound()})
		bfd.removePastOrInvalidRecords()
	}

	bfd.append(&headerInfo{
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  headerHash,
		state: state,
	})

	probableHighestNonce := bfd.computeProbableHighestNonce()
	bfd.setLastBlockRound(uint64(bfd.rounder.Index()))
	bfd.setProbableHighestNonce(probableHighestNonce)

	return nil
}

func (bfd *basicForkDetector) addFinalHeaders(finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) {
	if finalHeaders == nil {
		bfd.setFinalCheckpoint(bfd.lastCheckpoint())
		return
	}

	for i := 0; i < len(finalHeaders); i++ {
		isFinalHeaderNonceHigherThanCurrent := finalHeaders[i].GetNonce() > bfd.GetHighestFinalBlockNonce()
		if isFinalHeaderNonceHigherThanCurrent {
			if i == 0 {
				bfd.setFinalCheckpoint(&checkpointInfo{nonce: finalHeaders[i].GetNonce(), round: finalHeaders[i].GetRound()})
			}

			bfd.append(&headerInfo{
				nonce: finalHeaders[i].GetNonce(),
				round: finalHeaders[i].GetRound(),
				hash:  finalHeadersHashes[i],
				state: process.BHNotarized,
			})
		}
	}
}

func (bfd *basicForkDetector) removePastOrInvalidRecords() {
	bfd.removePastHeaders()
	bfd.removeInvalidHeaders()
	bfd.removePastCheckpoints()
}

func (bfd *basicForkDetector) checkBlockValidity(header data.HeaderHandler, state process.BlockHeaderState) error {
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

func (bfd *basicForkDetector) removePastHeaders() {
	finalCheckpointNonce := bfd.finalCheckpoint().nonce

	bfd.mutHeaders.Lock()
	for nonce := range bfd.headers {
		if nonce < finalCheckpointNonce {
			delete(bfd.headers, nonce)
		}
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *basicForkDetector) removeInvalidHeaders() {
	finalCheckpointRound := bfd.finalCheckpoint().round
	finalCheckpointNonce := bfd.finalCheckpoint().nonce

	var validHdrInfos []*headerInfo

	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		validHdrInfos = nil
		for i := 0; i < len(hdrInfos); i++ {
			roundDif := int64(hdrInfos[i].round) - int64(finalCheckpointRound)
			nonceDif := int64(hdrInfos[i].nonce) - int64(finalCheckpointNonce)
			if roundDif >= nonceDif {
				validHdrInfos = append(validHdrInfos, hdrInfos[i])
			}
		}
		if validHdrInfos == nil {
			delete(bfd.headers, nonce)
			continue
		}

		bfd.headers[nonce] = validHdrInfos
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *basicForkDetector) removePastCheckpoints() {
	bfd.removeCheckpointsBehindNonce(bfd.finalCheckpoint().nonce)
}

func (bfd *basicForkDetector) removeCheckpointsBehindNonce(nonce uint64) {
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
func (bfd *basicForkDetector) computeProbableHighestNonce() uint64 {
	probableHighestNonce := bfd.finalCheckpoint().nonce

	bfd.mutHeaders.RLock()
	for _, headersInfo := range bfd.headers {
		nonce := bfd.getProbableHighestNonce(headersInfo)
		if nonce <= probableHighestNonce {
			continue
		}
		probableHighestNonce = nonce
	}
	bfd.mutHeaders.RUnlock()

	return probableHighestNonce
}

func (bfd *basicForkDetector) getProbableHighestNonce(headersInfo []*headerInfo) uint64 {
	maxNonce := uint64(0)

	for _, headerInfo := range headersInfo {
		nonce := headerInfo.nonce
		// if header stored state is BHProposed, then the probable highest nonce should be set to its nonce-1, because
		// at that point the consensus was not achieved on this block and the only certainty is that the probable
		// highest nonce is nonce-1 on which this proposed block is constructed. This approach would avoid a situation
		// in which a proposed block on which the consensus would not be achieved would set all the nodes in sync mode,
		// because of the probable highest nonce set with its nonce instead of nonce-1.
		if headerInfo.state == process.BHProposed {
			nonce--
		}
		if nonce > maxNonce {
			maxNonce = nonce
		}
	}

	return maxNonce
}

// RemoveHeaders removes all the stored headers with a given nonce
func (bfd *basicForkDetector) RemoveHeaders(nonce uint64, hash []byte) {
	checkpointNonce := bfd.lastCheckpoint().nonce
	if nonce == checkpointNonce {
		bfd.removeCheckpointWithNonce(checkpointNonce)
	}

	var preservedHdrInfos []*headerInfo

	bfd.mutHeaders.RLock()
	hdrInfos := bfd.headers[nonce]
	bfd.mutHeaders.RUnlock()

	for _, hdrInfoStored := range hdrInfos {
		if bytes.Equal(hdrInfoStored.hash, hash) {
			continue
		}

		preservedHdrInfos = append(preservedHdrInfos, hdrInfoStored)
	}

	bfd.mutHeaders.Lock()
	if preservedHdrInfos == nil {
		delete(bfd.headers, nonce)
	} else {
		bfd.headers[nonce] = preservedHdrInfos
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *basicForkDetector) removeCheckpointWithNonce(nonce uint64) {
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
func (bfd *basicForkDetector) append(hdrInfo *headerInfo) {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

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
				if hdrInfo.state == process.BHNotarized {
					hdrInfoStored.state = process.BHNotarized
				} else if hdrInfo.state == process.BHProcessed {
					hdrInfoStored.state = process.BHProcessed
				}
			}
			return
		}
	}

	bfd.headers[hdrInfo.nonce] = append(bfd.headers[hdrInfo.nonce], hdrInfo)
}

// CheckFork method checks if the node could be on the fork
func (bfd *basicForkDetector) CheckFork() (bool, uint64, []byte) {
	var lowestForkNonce uint64
	var hashOfLowestForkNonce []byte
	var lowestRoundInForkNonce uint64
	var forkHeaderHash []byte
	var selfHdrInfo *headerInfo
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
			// Proposed blocks received do not count for fork choice, as they are not valid until the consensus
			// is achieved. They should be received afterwards through sync mechanism.
			if hdrInfos[i].state == process.BHProposed {
				continue
			}
			if hdrInfos[i].state == process.BHProcessed {
				selfHdrInfo = hdrInfos[i]
				continue
			}
			if hdrInfos[i].state == process.BHNotarized {
				if lowestRoundInForkNonce > 0 {
					lowestRoundInForkNonce = 0
					forkHeaderHash = hdrInfos[i].hash
					continue
				}

				if lowestRoundInForkNonce == 0 && bytes.Compare(hdrInfos[i].hash, forkHeaderHash) < 0 {
					forkHeaderHash = hdrInfos[i].hash
				}

				continue
			}

			if hdrInfos[i].round < lowestRoundInForkNonce {
				lowestRoundInForkNonce = hdrInfos[i].round
				forkHeaderHash = hdrInfos[i].hash
				continue
			}

			if hdrInfos[i].round == lowestRoundInForkNonce && bytes.Compare(hdrInfos[i].hash, forkHeaderHash) < 0 {
				forkHeaderHash = hdrInfos[i].hash
			}
		}

		if selfHdrInfo == nil {
			// if current nonce has not been processed yet, then skip and check the next one.
			continue
		}

		shouldSignalFork := selfHdrInfo.round > lowestRoundInForkNonce ||
			(selfHdrInfo.round == lowestRoundInForkNonce && strings.Compare(string(selfHdrInfo.hash), string(forkHeaderHash)) > 0)

		if !shouldSignalFork {
			// keep it clean so next time this position will be processed faster
			delete(bfd.headers, nonce)
			bfd.headers[nonce] = []*headerInfo{selfHdrInfo}
			continue
		}

		forkDetected = true
		if nonce < lowestForkNonce {
			lowestForkNonce = nonce
			hashOfLowestForkNonce = forkHeaderHash
		}
	}
	bfd.mutHeaders.Unlock()

	return forkDetected, lowestForkNonce, hashOfLowestForkNonce
}

// GetHighestFinalBlockNonce gets the highest nonce of the block which is final and it can not be reverted anymore
func (bfd *basicForkDetector) GetHighestFinalBlockNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

// ProbableHighestNonce gets the probable highest nonce
func (bfd *basicForkDetector) ProbableHighestNonce() uint64 {
	return bfd.probableHighestNonce()
}

// ResetProbableHighestNonceIfNeeded resets the probableHighestNonce to checkpoint if after maxRoundsToWait nothing
// is received so the node will act as synchronized
func (bfd *basicForkDetector) ResetProbableHighestNonceIfNeeded() {
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

func (bfd *basicForkDetector) addCheckpoint(checkpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	bfd.fork.checkpoint = append(bfd.fork.checkpoint, checkpoint)
	bfd.mutFork.Unlock()
}

func (bfd *basicForkDetector) lastCheckpoint() *checkpointInfo {
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

func (bfd *basicForkDetector) setFinalCheckpoint(finalCheckpoint *checkpointInfo) {
	bfd.mutFork.Lock()
	bfd.fork.finalCheckpoint = finalCheckpoint
	bfd.mutFork.Unlock()
}

func (bfd *basicForkDetector) finalCheckpoint() *checkpointInfo {
	bfd.mutFork.RLock()
	finalCheckpoint := bfd.fork.finalCheckpoint
	bfd.mutFork.RUnlock()

	return finalCheckpoint
}

func (bfd *basicForkDetector) setProbableHighestNonce(nonce uint64) {
	bfd.mutFork.Lock()
	bfd.fork.probableHighestNonce = nonce
	bfd.mutFork.Unlock()
}

func (bfd *basicForkDetector) probableHighestNonce() uint64 {
	bfd.mutFork.RLock()
	probableHighestNonce := bfd.fork.probableHighestNonce
	bfd.mutFork.RUnlock()

	return probableHighestNonce
}

func (bfd *basicForkDetector) setLastBlockRound(round uint64) {
	bfd.mutFork.Lock()
	bfd.fork.lastBlockRound = round
	bfd.mutFork.Unlock()
}

func (bfd *basicForkDetector) lastBlockRound() uint64 {
	bfd.mutFork.RLock()
	lastBlockRound := bfd.fork.lastBlockRound
	bfd.mutFork.RUnlock()

	return lastBlockRound
}

// IsInterfaceNil returns true if there is no value under the interface
func (bfd *basicForkDetector) IsInterfaceNil() bool {
	if bfd == nil {
		return true
	}
	return false
}
