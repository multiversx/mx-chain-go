package sync

import (
	"bytes"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type headerInfo struct {
	nonce       uint64
	hash        []byte
	isProcessed bool
	isSigned    bool
}

// basicForkDetector defines a struct with necessary data needed for fork detection
type basicForkDetector struct {
	headers             map[uint64][]*headerInfo
	mutHeaders          sync.RWMutex
	lastCheckpointNonce uint64
	checkpointNonce     uint64
}

// NewBasicForkDetector method creates a new BasicForkDetector object
func NewBasicForkDetector() *basicForkDetector {
	bfd := &basicForkDetector{}
	bfd.headers = make(map[uint64][]*headerInfo)
	return bfd
}

// AddHeader method adds a new header to headers map
func (bfd *basicForkDetector) AddHeader(header *block.Header, hash []byte, isProcessed bool) error {
	if header == nil {
		return ErrNilHeader
	}
	if hash == nil {
		return ErrNilHash
	}
	if header.Nonce < bfd.lastCheckpointNonce {
		return ErrLowerNonceInBlock
	}

	isSigned := isSigned(header)
	if isSigned && isProcessed {
		// create a check point and remove all the past headers
		bfd.lastCheckpointNonce = bfd.checkpointNonce
		bfd.checkpointNonce = header.Nonce
		bfd.removePastHeaders(bfd.lastCheckpointNonce)
	}

	bfd.append(&headerInfo{
		nonce:       header.Nonce,
		hash:        hash,
		isProcessed: isProcessed,
		isSigned:    isSigned,
	})

	return nil
}

func (bfd *basicForkDetector) removePastHeaders(nonce uint64) {
	bfd.mutHeaders.Lock()

	for storedNonce := range bfd.headers {
		if storedNonce < nonce {
			delete(bfd.headers, storedNonce)
		}
	}

	bfd.mutHeaders.Unlock()
}

// RemoveHeaders removes all the stored headers with a given nonce
func (bfd *basicForkDetector) RemoveHeaders(nonce uint64) {
	if nonce == bfd.checkpointNonce {
		bfd.checkpointNonce = bfd.lastCheckpointNonce
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
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	var lowestForkNonce uint64
	var selfHdrInfo *headerInfo
	lowestForkNonce = math.MaxUint64
	forkDetected := false

	for nonce, hdrInfos := range bfd.headers {
		if len(hdrInfos) == 1 {
			continue
		}

		selfHdrInfo = nil
		foundSignedBlock := false

		for i := 0; i < len(hdrInfos); i++ {
			if hdrInfos[i].isProcessed {
				selfHdrInfo = hdrInfos[i]
				continue
			}

			if hdrInfos[i].isSigned {
				foundSignedBlock = true
			}
		}

		if selfHdrInfo == nil {
			// if current nonce has not been processed yet, then skip and check the next one.
			continue
		}

		if selfHdrInfo.isSigned {
			// keep it clean so next time this position will be processed faster
			delete(bfd.headers, nonce)
			bfd.headers[nonce] = []*headerInfo{selfHdrInfo}
			continue
		}

		if foundSignedBlock {
			// fork detected: self has a processed unsigned block, and it received a signed block with the same nonce
			forkDetected = true
			if nonce < lowestForkNonce {
				lowestForkNonce = nonce
			}
		}
	}

	return forkDetected, lowestForkNonce
}

// GetHighestSignedBlockNonce gets the highest stored nonce of one signed block received/processed
func (bfd *basicForkDetector) GetHighestSignedBlockNonce() uint64 {
	bfd.mutHeaders.RLock()
	highestNonce := bfd.checkpointNonce
	for nonce, hdrInfos := range bfd.headers {
		if nonce <= highestNonce {
			continue
		}
		for i := 0; i < len(hdrInfos); i++ {
			if hdrInfos[i].isSigned {
				highestNonce = nonce
				break
			}
		}
	}
	bfd.mutHeaders.RUnlock()
	return highestNonce
}

// GetHighestFinalBlockNonce gets the highest nonce of the block which is final and it can not be reverted anymore
func (bfd *basicForkDetector) GetHighestFinalBlockNonce() uint64 {
	return bfd.lastCheckpointNonce
}
