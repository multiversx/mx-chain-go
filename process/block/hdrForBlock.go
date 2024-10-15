package block

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
)

type hdrForBlock struct {
	missingHdrs                  uint32
	missingFinalityAttestingHdrs uint32
	highestHdrNonce              map[uint32]uint64
	mutHdrsForBlock              sync.RWMutex
	hdrHashAndInfo               map[string]*hdrInfo
}

func newHdrForBlock() *hdrForBlock {
	return &hdrForBlock{
		hdrHashAndInfo:  make(map[string]*hdrInfo),
		highestHdrNonce: make(map[uint32]uint64),
	}
}

func (hfb *hdrForBlock) initMaps() {
	hfb.mutHdrsForBlock.Lock()
	hfb.hdrHashAndInfo = make(map[string]*hdrInfo)
	hfb.highestHdrNonce = make(map[uint32]uint64)
	hfb.mutHdrsForBlock.Unlock()
}

func (hfb *hdrForBlock) resetMissingHdrs() {
	hfb.mutHdrsForBlock.Lock()
	hfb.missingHdrs = 0
	hfb.missingFinalityAttestingHdrs = 0
	hfb.mutHdrsForBlock.Unlock()
}

func (hfb *hdrForBlock) getHdrHashMap() map[string]data.HeaderHandler {
	m := make(map[string]data.HeaderHandler)

	hfb.mutHdrsForBlock.RLock()
	for hash, hi := range hfb.hdrHashAndInfo {
		m[hash] = hi.hdr
	}
	hfb.mutHdrsForBlock.RUnlock()

	return m
}
