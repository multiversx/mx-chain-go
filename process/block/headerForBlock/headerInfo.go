package headerForBlock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type headerInfo struct {
	sync.RWMutex
	hdr               data.HeaderHandler
	usedInBlock       bool
	hasProof          bool
	hasProofRequested bool
}

// NewHeaderInfo returns a new instance of headerInfo
func NewHeaderInfo(
	hdr data.HeaderHandler,
	usedInBlock bool,
	hasProof bool,
	hasProofRequested bool,
) *headerInfo {
	return &headerInfo{
		hdr:               hdr,
		usedInBlock:       usedInBlock,
		hasProof:          hasProof,
		hasProofRequested: hasProofRequested,
	}
}

// NewEmptyHeaderInfo returns a new instance of headerInfo with nothing set
func NewEmptyHeaderInfo() *headerInfo {
	return &headerInfo{}
}

// GetHeader returns the header
func (hi *headerInfo) GetHeader() data.HeaderHandler {
	hi.RLock()
	defer hi.RUnlock()

	return hi.hdr
}

// UsedInBlock returns the usedInBlock field
func (hi *headerInfo) UsedInBlock() bool {
	hi.RLock()
	defer hi.RUnlock()

	return hi.usedInBlock
}

// HasProof returns the hasProof field
func (hi *headerInfo) HasProof() bool {
	hi.RLock()
	defer hi.RUnlock()

	return hi.hasProof
}

// HasProofRequested returns the hasProofRequested field
func (hi *headerInfo) HasProofRequested() bool {
	hi.RLock()
	defer hi.RUnlock()

	return hi.hasProofRequested
}

// SetUsedInBlock sets the usedInBlock field
func (hi *headerInfo) SetUsedInBlock(usedInBlock bool) {
	hi.Lock()
	defer hi.Unlock()

	hi.usedInBlock = usedInBlock
}

// SetHeader sets the header
func (hi *headerInfo) SetHeader(hdr data.HeaderHandler) {
	if check.IfNil(hdr) {
		return
	}

	hi.Lock()
	defer hi.Unlock()

	hi.hdr = hdr
}

// SetHasProof sets the hasProof field
func (hi *headerInfo) SetHasProof(hasProof bool) {
	hi.Lock()
	defer hi.Unlock()

	hi.hasProof = hasProof
}

// SetHasProofRequested sets the hasProofRequested field
func (hi *headerInfo) SetHasProofRequested(hasProofRequested bool) {
	hi.Lock()
	defer hi.Unlock()

	hi.hasProofRequested = hasProofRequested
}

// IsInterfaceNil returns true if there is no value under the interface
func (hi *headerInfo) IsInterfaceNil() bool {
	return hi == nil
}
