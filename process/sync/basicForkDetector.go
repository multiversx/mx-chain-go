package sync

import (
	"bytes"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type headerInfo struct {
	nonce uint64
	round uint32
	hash  []byte
	state process.BlockHeaderState
}

type forkInfo struct {
	checkpointNonce      uint64
	lastCheckpointNonce  uint64
	checkpointRound      int32
	lastCheckpointRound  int32
	probableHighestNonce uint64
	lastBlockRound       int32
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
	if rounder == nil {
		return nil, process.ErrNilRounder
	}

	bfd := &basicForkDetector{
		rounder: rounder,
	}

	bfd.fork.checkpointRound = -1
	bfd.fork.lastCheckpointRound = -1
	bfd.fork.lastBlockRound = -1
	bfd.headers = make(map[uint64][]*headerInfo)

	return bfd, nil
}

// AddHeader method adds a new header to headers map
func (bfd *basicForkDetector) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
	if header == nil {
		return ErrNilHeader
	}
	if hash == nil {
		return ErrNilHash
	}

	err := bfd.checkBlockValidity(header, state)
	if err != nil {
		return err
	}

	if state == process.BHProcessed {
		// create a check point and remove all the past headers
		bfd.mutFork.Lock()
		bfd.fork.lastCheckpointNonce = bfd.fork.checkpointNonce
		bfd.fork.lastCheckpointRound = bfd.fork.checkpointRound
		bfd.fork.checkpointNonce = header.GetNonce()
		bfd.fork.checkpointRound = int32(header.GetRound())
		bfd.mutFork.Unlock()
		bfd.removePastHeaders()
		bfd.removeInvalidHeaders()
	}

	bfd.append(&headerInfo{
		nonce: header.GetNonce(),
		round: header.GetRound(),
		hash:  hash,
		state: state,
	})

	probableHighestNonce := bfd.computeProbableHighestNonce()

	bfd.mutFork.RLock()
	bfd.fork.lastBlockRound = bfd.rounder.Index()
	bfd.fork.probableHighestNonce = probableHighestNonce
	bfd.mutFork.RUnlock()

	return nil
}

func (bfd *basicForkDetector) checkBlockValidity(header data.HeaderHandler, state process.BlockHeaderState) error {
	bfd.mutFork.RLock()
	roundDif := int32(header.GetRound()) - bfd.fork.lastCheckpointRound
	nonceDif := int64(header.GetNonce() - bfd.fork.lastCheckpointNonce)
	bfd.mutFork.RUnlock()

	if roundDif < 0 {
		return ErrLowerRoundInBlock
	}
	if nonceDif < 0 {
		return ErrLowerNonceInBlock
	}
	if int32(header.GetRound()) > bfd.rounder.Index() {
		return ErrHigherRoundInBlock
	}
	if int64(roundDif) < nonceDif {
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
	bfd.mutFork.RLock()
	lastCheckpointNonce := bfd.fork.lastCheckpointNonce
	bfd.mutFork.RUnlock()

	bfd.mutHeaders.Lock()
	for nonce := range bfd.headers {
		if nonce < lastCheckpointNonce {
			delete(bfd.headers, nonce)
		}
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *basicForkDetector) removeInvalidHeaders() {
	bfd.mutFork.RLock()
	lastCheckpointRound := bfd.fork.lastCheckpointRound
	lastCheckpointNonce := bfd.fork.lastCheckpointNonce
	bfd.mutFork.RUnlock()

	var validHdrInfos []*headerInfo

	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		validHdrInfos = nil
		for i := 0; i < len(hdrInfos); i++ {
			roundDif := int32(hdrInfos[i].round) - lastCheckpointRound
			nonceDif := int64(hdrInfos[i].nonce - lastCheckpointNonce)
			if int64(roundDif) >= nonceDif {
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

// computeProbableHighestNonce computes the probable highest nonce from the valid received/processed headers
func (bfd *basicForkDetector) computeProbableHighestNonce() uint64 {
	bfd.mutFork.RLock()
	probableHighestNonce := bfd.fork.lastCheckpointNonce
	bfd.mutFork.RUnlock()

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
	bfd.mutFork.Lock()
	if nonce == bfd.fork.checkpointNonce {
		bfd.fork.checkpointNonce = bfd.fork.lastCheckpointNonce
		bfd.fork.checkpointRound = bfd.fork.lastCheckpointRound
	}
	bfd.mutFork.Unlock()

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
			if hdrInfoStored.state != process.BHProcessed && hdrInfo.state == process.BHProcessed {
				// if the stored and received headers processed at the same time have equal hashes, that the old record
				// will be replaced with the processed one. This nonce is marked at bootsrapping as processed, but as it
				// is also received through broadcasting, the system stores as received.
				hdrInfoStored.state = process.BHProcessed
			}
			return
		}
	}

	bfd.headers[hdrInfo.nonce] = append(bfd.headers[hdrInfo.nonce], hdrInfo)
}

// CheckFork method checks if the node could be on the fork
func (bfd *basicForkDetector) CheckFork() (bool, uint64) {
	var lowestForkNonce uint64
	var lowestRoundInForkNonce uint32
	var selfHdrInfo *headerInfo
	lowestForkNonce = math.MaxUint64
	forkDetected := false

	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		if len(hdrInfos) == 1 {
			continue
		}

		selfHdrInfo = nil
		lowestRoundInForkNonce = math.MaxUint32

		for i := 0; i < len(hdrInfos); i++ {
			if hdrInfos[i].state == process.BHProcessed {
				selfHdrInfo = hdrInfos[i]
				continue
			}

			if hdrInfos[i].round < lowestRoundInForkNonce {
				lowestRoundInForkNonce = hdrInfos[i].round
			}
		}

		if selfHdrInfo == nil {
			// if current nonce has not been processed yet, then skip and check the next one.
			continue
		}

		if selfHdrInfo.round <= lowestRoundInForkNonce {
			// keep it clean so next time this position will be processed faster
			delete(bfd.headers, nonce)
			bfd.headers[nonce] = []*headerInfo{selfHdrInfo}
			continue
		}

		forkDetected = true
		if nonce < lowestForkNonce {
			lowestForkNonce = nonce
		}
	}
	bfd.mutHeaders.Unlock()

	return forkDetected, lowestForkNonce
}

// GetHighestFinalBlockNonce gets the highest nonce of the block which is final and it can not be reverted anymore
func (bfd *basicForkDetector) GetHighestFinalBlockNonce() uint64 {
	bfd.mutFork.RLock()
	lastCheckpointNonce := bfd.fork.lastCheckpointNonce
	bfd.mutFork.RUnlock()

	return lastCheckpointNonce
}

// ProbableHighestNonce gets the probable highest nonce
func (bfd *basicForkDetector) ProbableHighestNonce() uint64 {
	// TODO: This fallback mechanism should be improved
	// This mechanism is necessary to manage the case when the node will act as synchronized because no new block,
	// higher than its checkpoint, would be received anymore (this could be the case when during an epoch, the number of
	// validators in one shard decrease under the size of 2/3 + 1 of the consensus group. In this case no new block would
	// be proposed anymore and any node which would try to boostrap, would be stuck at the genesis block. This case could
	// be solved, if the proposed blocks received from leaders would also call the AddHeader method of this class).

	// If after maxRoundsToWait nothing is received, the probableHighestNonce will be set to checkpoint,
	// so the node will act as synchronized
	bfd.mutFork.Lock()
	roundsWithoutReceivedBlock := bfd.rounder.Index() - bfd.fork.lastBlockRound
	if roundsWithoutReceivedBlock > maxRoundsToWait {
		if bfd.fork.probableHighestNonce > bfd.fork.checkpointNonce {
			bfd.fork.probableHighestNonce = bfd.fork.checkpointNonce
		}
	}
	probableHighestNonce := bfd.fork.probableHighestNonce
	bfd.mutFork.Unlock()

	return probableHighestNonce
}
