package headerForBlock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
)

type lastNotarizedHeaderInfo struct {
	sync.RWMutex
	header                data.HeaderHandler
	hash                  []byte
	notarizedBasedOnProof bool
	hasProof              bool
}

// newLastNotarizedHeaderInfo returns a new instance of lastNotarizedHeaderInfo
func newLastNotarizedHeaderInfo(
	header data.HeaderHandler,
	hash []byte,
	notarizedBasedOnProof bool,
	hasProof bool,
) *lastNotarizedHeaderInfo {
	return &lastNotarizedHeaderInfo{
		header:                header,
		hash:                  hash,
		notarizedBasedOnProof: notarizedBasedOnProof,
		hasProof:              hasProof,
	}
}

// GetHeader returns the header
func (lnhi *lastNotarizedHeaderInfo) GetHeader() data.HeaderHandler {
	lnhi.RLock()
	defer lnhi.RUnlock()

	return lnhi.header
}

// GetHash returns the header hash
func (lnhi *lastNotarizedHeaderInfo) GetHash() []byte {
	lnhi.RLock()
	defer lnhi.RUnlock()

	return lnhi.hash
}

// SetHeader sets the header
func (lnhi *lastNotarizedHeaderInfo) SetHeader(hdr data.HeaderHandler) {
	lnhi.Lock()
	lnhi.header = hdr
	lnhi.Unlock()
}

// SetHash sets the header hash
func (lnhi *lastNotarizedHeaderInfo) SetHash(hash []byte) {
	lnhi.Lock()
	lnhi.hash = hash
	lnhi.Unlock()
}

// HasProof returns the hasProof field
func (lnhi *lastNotarizedHeaderInfo) HasProof() bool {
	lnhi.RLock()
	defer lnhi.RUnlock()

	return lnhi.hasProof
}

// SetHasProof sets the hasProof field
func (lnhi *lastNotarizedHeaderInfo) SetHasProof(hasProof bool) {
	lnhi.Lock()
	lnhi.hasProof = hasProof
	lnhi.Unlock()
}

// NotarizedBasedOnProof returns the notarizedBasedOnProof field
func (lnhi *lastNotarizedHeaderInfo) NotarizedBasedOnProof() bool {
	lnhi.RLock()
	defer lnhi.RUnlock()

	return lnhi.notarizedBasedOnProof
}

// SetNotarizedBasedOnProof sets the notarizedBasedOnProof field
func (lnhi *lastNotarizedHeaderInfo) SetNotarizedBasedOnProof(notarizedBasedOnProof bool) {
	lnhi.Lock()
	lnhi.notarizedBasedOnProof = notarizedBasedOnProof
	lnhi.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (lnhi *lastNotarizedHeaderInfo) IsInterfaceNil() bool {
	return lnhi == nil
}
