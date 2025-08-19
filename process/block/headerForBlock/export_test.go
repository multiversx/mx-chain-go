package headerForBlock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

// NewHeaderInfo -
func NewHeaderInfo(
	hdr data.HeaderHandler,
	usedInBlock bool,
	hasProof bool,
	hasProofRequested bool,
) *headerInfo {
	return newHeaderInfo(hdr, usedInBlock, hasProof, hasProofRequested)
}

// NewEmptyHeaderInfo -
func NewEmptyHeaderInfo() *headerInfo {
	return newEmptyHeaderInfo()
}

// NewLastNotarizedHeaderInfo -
func NewLastNotarizedHeaderInfo(
	header data.HeaderHandler,
	hash []byte,
	notarizedBasedOnProof bool,
	hasProof bool,
) *lastNotarizedHeaderInfo {
	return newLastNotarizedHeaderInfo(header, hash, notarizedBasedOnProof, hasProof)
}

// SetRequestHandler -
func (hfb *headersForBlock) SetRequestHandler(handler process.RequestHandler) {
	hfb.requestHandler = handler
}
