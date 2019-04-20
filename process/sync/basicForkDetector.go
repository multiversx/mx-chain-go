package sync

import (
	"bytes"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type headerInfo struct {
	nonce       uint64
	round       uint32
	hash        []byte
	isProcessed bool
}

// basicForkDetector defines a struct with necessary data needed for fork detection
type basicForkDetector struct {
	rounder consensus.Rounder

	headers              map[uint64][]*headerInfo
	mutHeaders           sync.RWMutex
	lastCheckpointNonce  uint64
	checkpointNonce      uint64
	lastCheckpointRound  int32
	checkpointRound      int32
	probableHighestNonce uint64
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

	bfd.checkpointRound = -1
	bfd.lastCheckpointRound = -1
	bfd.headers = make(map[uint64][]*headerInfo)
	return bfd, nil
}

// AddHeader method adds a new header to headers map
func (bfd *basicForkDetector) AddHeader(header data.HeaderHandler, hash []byte, isProcessed bool) error {
	if header == nil {
		return ErrNilHeader
	}
	if hash == nil {
		return ErrNilHash
	}

	err := bfd.checkBlockValidity(header)
	if err != nil {
		return err
	}

	if isProcessed {
		// create a check point and remove all the past headers
		bfd.lastCheckpointNonce = bfd.checkpointNonce
		bfd.lastCheckpointRound = bfd.checkpointRound
		bfd.checkpointNonce = header.GetNonce()
		bfd.checkpointRound = int32(header.GetRound())
		bfd.removePastHeaders()
		bfd.removeInvalidHeaders()
	}

	bfd.append(&headerInfo{
		nonce:       header.GetNonce(),
		round:       header.GetRound(),
		hash:        hash,
		isProcessed: isProcessed,
	})

	bfd.probableHighestNonce = bfd.computeProbableHighestNonce()

	return nil
}

func (bfd *basicForkDetector) checkBlockValidity(header data.HeaderHandler) error {
	roundDif := int32(header.GetRound()) - bfd.lastCheckpointRound
	nonceDif := int64(header.GetNonce() - bfd.lastCheckpointNonce)

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
	if !isSigned(header) {
		return ErrBlockIsNotSigned
	}

	return nil
}

func (bfd *basicForkDetector) removePastHeaders() {
	bfd.mutHeaders.Lock()
	for nonce := range bfd.headers {
		if nonce < bfd.lastCheckpointNonce {
			delete(bfd.headers, nonce)
		}
	}
	bfd.mutHeaders.Unlock()
}

func (bfd *basicForkDetector) removeInvalidHeaders() {
	var validHdrInfos []*headerInfo
	bfd.mutHeaders.Lock()
	for nonce, hdrInfos := range bfd.headers {
		validHdrInfos = nil
		for i := 0; i < len(hdrInfos); i++ {
			roundDif := int32(hdrInfos[i].round) - bfd.lastCheckpointRound
			nonceDif := int64(hdrInfos[i].nonce - bfd.lastCheckpointNonce)
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
	probableHighestNonce := bfd.lastCheckpointNonce
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
func (bfd *basicForkDetector) RemoveHeaders(nonce uint64) {
	if nonce == bfd.checkpointNonce {
		bfd.checkpointNonce = bfd.lastCheckpointNonce
		bfd.checkpointRound = bfd.lastCheckpointRound
	}

	bfd.mutHeaders.Lock()
	delete(bfd.headers, nonce)
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
			if !hdrInfoStored.isProcessed && hdrInfo.isProcessed {
				// if the stored and received headers processed at the same time have equal hashes, that the old record
				// will be replaced with the processed one. This nonce is marked at bootsrapping as processed, but as it
				// is also received through broadcasting, the system stores as received.
				hdrInfoStored.isProcessed = true
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
			if hdrInfos[i].isProcessed {
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
	return bfd.lastCheckpointNonce
}

// ProbableHighestNonce gets the probable highest nonce
func (bfd *basicForkDetector) ProbableHighestNonce() uint64 {
	return bfd.probableHighestNonce
}

// ResetProbableHighestNonce sets the probable highest nonce to the value of checkpoint nonce
func (bfd *basicForkDetector) ResetProbableHighestNonce() {
	bfd.probableHighestNonce = bfd.checkpointNonce
}
