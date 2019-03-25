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
	mutHeaders      sync.Mutex
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
	if header.Nonce < bfd.checkpointNonce {
		return ErrLowerNonceInBlock
	}

	if !isEmpty(header) && isProcessed {
		// create a check point and remove all the past headers
		bfd.checkpointNonce = header.Nonce
		bfd.removePastHeaders(header.Nonce)
	}

	bfd.append(&headerInfo{
		header:      header,
		hash:        hash,
		isProcessed: isProcessed,
	})

	return nil
}

func (bfd *basicForkDetector) removePastHeaders(nonce uint64) {
	bfd.mutHeaders.Lock()

	for storedNonce := range bfd.headers {
		if storedNonce <= nonce {
			delete(bfd.headers, nonce)
		}
	}

	bfd.mutHeaders.Unlock()
}

// RemoveProcessedHeader removes all stored headers with a given nonce
func (bfd *basicForkDetector) RemoveProcessedHeader(nonce uint64) error {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	hdrInfosStored := bfd.headers[nonce]
	isHdrInfosStoredNilOrEmpty := hdrInfosStored == nil || len(hdrInfosStored) == 0
	if isHdrInfosStoredNilOrEmpty {
		return ErrNilOrEmptyInfoStored
	}

	var newHdrInfosStored []*headerInfo
	for _, hdrInfoStored := range hdrInfosStored {
		if !hdrInfoStored.isProcessed {
			newHdrInfosStored = append(newHdrInfosStored, hdrInfoStored)
		}
	}

	isNewHdrInfosStoredNilOrEmpty := newHdrInfosStored == nil || len(newHdrInfosStored) == 0
	if isNewHdrInfosStoredNilOrEmpty {
		delete(bfd.headers, nonce)
	} else {
		bfd.headers[nonce] = newHdrInfosStored
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
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	var selfHdrInfo *headerInfo
	for nonce, hdrInfos := range bfd.headers {
		if len(hdrInfos) == 1 {
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

		if selfHdrInfo == nil {
			//current nonce has not been (yet) processed, skipping, trying the next one
			continue
		}

		if !isEmpty(selfHdrInfo.header) {
			//keep it clean so next time this position will be processed faster
			delete(bfd.headers, nonce)
			bfd.headers[nonce] = []*headerInfo{selfHdrInfo}
			continue
		}

		if foundNotEmptyBlock {
			//detected a fork: self has an unsigned header, it also received a signed block
			//with the same nonce
			return true
		}
	}

	return false
}
