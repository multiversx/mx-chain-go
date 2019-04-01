package sync

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type headerInfo struct {
	header      *block.Header
	hash        []byte
	isProcessed bool
}

// basicForkDetector defines a struct with necessary data needed for fork detection
type basicForkDetector struct {
	headers         map[uint64][]*headerInfo
	mutHeaders      sync.RWMutex
	checkpointNonce uint64
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

	if !isEmpty(header) && isProcessed {
		// create a check point
		bfd.checkpointNonce = header.Nonce
	}

	bfd.append(&headerInfo{
		header:      header,
		hash:        hash,
		isProcessed: isProcessed,
	})

	return nil
}

// ResetProcessedHeader removes all stored headers with a given nonce
func (bfd *basicForkDetector) ResetProcessedHeader(nonce uint64) error {
	bfd.mutHeaders.RLock()
	defer bfd.mutHeaders.RUnlock()

	hdrInfosStored := bfd.headers[nonce]
	isHdrInfosStoredNilOrEmpty := hdrInfosStored == nil || len(hdrInfosStored) == 0
	if isHdrInfosStoredNilOrEmpty {
		return ErrNilOrEmptyInfoStored
	}

	for _, hdrInfoStored := range hdrInfosStored {
		if hdrInfoStored.isProcessed {
			hdrInfoStored.isProcessed = false
			if !isEmpty(hdrInfoStored.header) {
				bfd.checkpointNonce = bfd.getLastCheckpointNonce()
			}
			break
		}
	}

	return nil
}

// append adds a new header in the slice found in nonce position
// it not adds the header if its hash is already stored in the slice
func (bfd *basicForkDetector) append(hdrInfo *headerInfo) {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	hdrInfos := bfd.headers[hdrInfo.header.Nonce]
	isHdrInfosNilOrEmpty := hdrInfos == nil || len(hdrInfos) == 0
	if isHdrInfosNilOrEmpty {
		bfd.headers[hdrInfo.header.Nonce] = []*headerInfo{hdrInfo}
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

	bfd.headers[hdrInfo.header.Nonce] = append(bfd.headers[hdrInfo.header.Nonce], hdrInfo)
}

// CheckFork method checks if the node could be on the fork
func (bfd *basicForkDetector) CheckFork() bool {
	bfd.mutHeaders.RLock()
	defer bfd.mutHeaders.RUnlock()

	var selfHdrInfo *headerInfo
	for nonce, hdrInfos := range bfd.headers {
		if len(hdrInfos) == 1 {
			continue
		}
		if nonce < bfd.checkpointNonce {
			continue
		}

		selfHdrInfo = nil
		foundNotEmptyBlock := false

		for i := 0; i < len(hdrInfos); i++ {
			if hdrInfos[i].isProcessed {
				selfHdrInfo = hdrInfos[i]
				continue
			}

			if !isEmpty(hdrInfos[i].header) {
				foundNotEmptyBlock = true
			}
		}

		if selfHdrInfo == nil || !isEmpty(selfHdrInfo.header) {
			// if current nonce has not been processed yet or it is processed and signed, then skipping and checking the next one
			continue
		}

		if foundNotEmptyBlock {
			// detected a fork: self has an unsigned header, it also received a signed block
			// with the same nonce
			return true
		}
	}

	return false
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
			if !isEmpty(hdrInfos[i].header) {
				highestNonce = nonce
				break
			}
		}
	}

	bfd.mutHeaders.RUnlock()

	return highestNonce
}

// getLastCheckpointNonce gets the last set checkpoint behind the current one
func (bfd *basicForkDetector) getLastCheckpointNonce() uint64 {
	lastCheckpointNonce := uint64(0)
	for nonce, hdrInfos := range bfd.headers {
		if nonce >= bfd.checkpointNonce || nonce <= lastCheckpointNonce {
			continue
		}
		for i := 0; i < len(hdrInfos); i++ {
			if !isEmpty(hdrInfos[i].header) && hdrInfos[i].isProcessed {
				lastCheckpointNonce = nonce
				break
			}
		}
	}
	return lastCheckpointNonce
}
